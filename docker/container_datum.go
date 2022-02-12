package docker

import (
	"dc-top/utils"
	"encoding/json"
	"log"
	"strings"

	"github.com/docker/docker/api/types"
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
		log.Println(err)
		if strings.HasPrefix(err.Error(), "invalid character") || isBeingRemoved(base.ID) || isDeleted(base.ID) {
			return ContainerDatum{
				base:         base,
				stats_stream: stats_stream,
				cached_stats: ContainerMainStats{},
				is_deleted:   true,
			}
		}
		log.Fatalf("1 Failed to get stats and container %s wasnt deleted", base.ID)
	}
	return ContainerDatum{
		base:         base,
		stats_stream: stats_stream,
		cached_stats: _cached_stats,
		is_deleted:   false,
	}
}

func UpdatedDatum(old_datum ContainerDatum) (ContainerDatum, error) {
	new_stats, err := getNewStats(old_datum.stats_stream)
	if err != nil {
		log.Println("2 Failed to get new container stats:")
		log.Println(err)
		if strings.HasPrefix(err.Error(), "invalid character") || isBeingRemoved(old_datum.base.ID) || isDeleted(old_datum.base.ID) {
			return ContainerDatum{
				base:         old_datum.base,
				stats_stream: old_datum.stats_stream,
				cached_stats: old_datum.cached_stats,
				is_deleted:   true,
			}, err
		}
		log.Fatalf("2 Failed to get stats and container %s wasn't deleted", old_datum.base.ID)
	}
	return ContainerDatum{
		base:         old_datum.base,
		stats_stream: old_datum.stats_stream,
		cached_stats: new_stats,
		is_deleted:   false,
	}, err
}

func (datum *ContainerDatum) ID() string {
	return datum.base.ID
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
	return container_stats_data, nil
}
