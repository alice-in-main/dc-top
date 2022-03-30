package docker

import (
	"context"
	"dc-top/utils"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
)

type ContainerDatum struct {
	base         types.Container
	stats_stream types.ContainerStats
	cached_stats ContainerMainStats
	inspection   types.ContainerJSON
	is_deleted   bool
}

func NewContainerDatum(ctx context.Context, base types.Container, stats_stream types.ContainerStats) (ContainerDatum, error) {
	_cached_stats, err := getNewStats(base.ID, &stats_stream)
	if err != nil {
		log.Println("1 Failed to get new container stats:")
		is_being_removed, test_err := isBeingRemoved(ctx, base.ID)
		if test_err != nil {
			return ContainerDatum{}, err
		}
		is_deleted, test_err := isDeleted(ctx, base.ID)
		if test_err != nil {
			return ContainerDatum{}, err
		}
		if strings.HasPrefix(err.Error(), "invalid character") || err == io.EOF || is_being_removed || is_deleted {
			return ContainerDatum{
				base:         base,
				stats_stream: stats_stream,
				cached_stats: ContainerMainStats{},
				inspection:   types.ContainerJSON{},
				is_deleted:   true,
			}, nil
		}
		return ContainerDatum{}, fmt.Errorf("1 Failed to get stats and container %s wasnt deleted. %s", base.ID, err)
	}
	return ContainerDatum{
		base:         base,
		stats_stream: stats_stream,
		cached_stats: _cached_stats,
		inspection:   InspectContainerNoPanic(ctx, base.ID),
		is_deleted:   false,
	}, nil
}

func UpdatedDatum(ctx context.Context, old_datum ContainerDatum, base types.Container) (ContainerDatum, error) {
	new_stats, err := getNewStatsWithPrev(&old_datum)
	if err != nil {
		log.Println("2 Failed to get new container stats:")
		is_being_removed, test_err := isBeingRemoved(ctx, old_datum.base.ID)
		if test_err != nil {
			return ContainerDatum{}, err
		}
		is_deleted, test_err := isDeleted(ctx, old_datum.base.ID)
		if test_err != nil {
			return ContainerDatum{}, err
		}
		if strings.HasPrefix(err.Error(), "unexpected end of JSON input") ||
			strings.HasPrefix(err.Error(), "invalid character") ||
			err == io.EOF || is_being_removed || is_deleted {
			return ContainerDatum{
				base:         old_datum.base,
				stats_stream: old_datum.stats_stream,
				cached_stats: old_datum.cached_stats,
				inspection:   old_datum.inspection,
				is_deleted:   true,
			}, nil
		}
		return ContainerDatum{}, fmt.Errorf("2 Failed to get stats and container %s wasnt deleted. %s", old_datum.base.ID, err)
	}
	return ContainerDatum{
		base:         base,
		stats_stream: old_datum.stats_stream,
		cached_stats: new_stats,
		inspection:   old_datum.inspection,
		is_deleted:   false,
	}, nil
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

func (datum *ContainerDatum) InspectData() types.ContainerJSON {
	return datum.inspection
}

func (datum *ContainerDatum) Contains(substr string) bool {
	return strings.Contains(datum.Image(), substr) || strings.Contains(datum.cached_stats.Name, substr)
}

func getNewStatsWithPrev(old_datum *ContainerDatum) (ContainerMainStats, error) {
	new_stats, err := getNewStats(old_datum.base.ID, &old_datum.stats_stream)
	new_stats.PreNetwork = old_datum.cached_stats.Network
	return new_stats, err
}

func getNewStats(id string, stats_stream *types.ContainerStats) (ContainerMainStats, error) {
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
	container_stats_data.Name = strings.TrimPrefix(container_stats_data.Name, "/")
	return container_stats_data, nil
}
