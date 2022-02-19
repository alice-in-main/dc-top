package docker

import (
	"context"
	"dc-top/utils"
	"encoding/json"
	"io"
	"log"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
)

type ContainerDatum struct {
	base         types.Container
	stats_stream types.ContainerStats
	cached_stats ContainerMainStats
	is_deleted   bool
}

func NewContainerDatum(base types.Container, stats_stream types.ContainerStats) ContainerDatum {
	_cached_stats, err := getNewStats(stats_stream)
	if err != nil {
		log.Println("1 Failed to get new container stats:")
		if strings.HasPrefix(err.Error(), "invalid character") || err == io.EOF || isBeingRemoved(base.ID) || isDeleted(base.ID) {
			return ContainerDatum{
				base:         base,
				stats_stream: stats_stream,
				cached_stats: ContainerMainStats{},
				is_deleted:   true,
			}
		}
		log.Fatalf("1 Failed to get stats and container %s wasnt deleted. %s", base.ID, err)
	}
	return ContainerDatum{
		base:         base,
		stats_stream: stats_stream,
		cached_stats: _cached_stats,
		is_deleted:   false,
	}
}

func UpdatedDatum(old_datum ContainerDatum) (ContainerDatum, error) {
	new_stats, err := getNewStatsWithPrev(old_datum)
	if err != nil {
		log.Println("2 Failed to get new container stats:")
		if strings.HasPrefix(err.Error(), "unexpected end of JSON input") ||
			strings.HasPrefix(err.Error(), "invalid character") ||
			err == io.EOF || isBeingRemoved(old_datum.base.ID) ||
			isDeleted(old_datum.base.ID) {
			return ContainerDatum{
				base:         old_datum.base,
				stats_stream: old_datum.stats_stream,
				cached_stats: old_datum.cached_stats,
				is_deleted:   true,
			}, err
		}
		log.Fatalf("2 Failed to get stats and container %s wasn't deleted. %s", old_datum.base.ID, err)
	}
	var filters filters.Args = filters.NewArgs(filters.Arg("id", old_datum.ID()))
	containers, err := docker_cli.ContainerList(context.Background(), types.ContainerListOptions{All: true, Filters: filters})
	if len(containers) != 1 {
		log.Println(containers)
		log.Fatal("Got more than 1 filtered image from id")
	}
	return ContainerDatum{
		base:         containers[0],
		stats_stream: old_datum.stats_stream,
		cached_stats: new_stats,
		is_deleted:   false,
	}, err
}

func (datum *ContainerDatum) ID() string {
	return datum.base.ID
}

func (datum *ContainerDatum) State() string {
	return datum.base.State
}

func (datum *ContainerDatum) Status() string {
	return datum.base.Status
}

func (datum *ContainerDatum) Image() string {
	return datum.base.Image
}

func (datum *ContainerDatum) CachedStats() ContainerMainStats {
	return datum.cached_stats
}

func (datum *ContainerDatum) Close() {
	datum.stats_stream.Body.Close()
}

func (datum *ContainerDatum) IsDeleted() bool {
	return datum.is_deleted
}

func getNewStatsWithPrev(old_datum ContainerDatum) (ContainerMainStats, error) {
	new_stats, err := getNewStats(old_datum.stats_stream)
	new_stats.PreNetwork = old_datum.cached_stats.Network
	return new_stats, err
}

func getNewStats(stats_stream types.ContainerStats) (ContainerMainStats, error) {
	const max_container_stats_data_len = 1 << 14
	container_stats_json_data := make([]byte, max_container_stats_data_len)
	_, err := stats_stream.Body.Read(container_stats_json_data)
	if err != nil {
		log.Printf("Got error '%s' while fetching container stats\n", err)
		return ContainerMainStats{}, err
	}
	nl_index := utils.FindByte('\n', container_stats_json_data)
	data_json_stats := container_stats_json_data[:nl_index]
	var container_stats_data ContainerMainStats
	err = json.Unmarshal(data_json_stats, &container_stats_data)
	if err != nil {
		log.Printf("Got error '%s' while fetching container stats\n", err)
		return ContainerMainStats{}, err
	}
	for key, network := range container_stats_data.Network {
		new_network := network
		new_network.LastUpdateTime = time.Now()
		container_stats_data.Network[key] = new_network
	}
	return container_stats_data, nil
}
