package docker

import (
	"dc-top/utils"
	"encoding/json"
	"io"
	"log"

	"github.com/docker/docker/api/types"
)

type ContainerDatum struct {
	base         types.Container
	stats_stream types.ContainerStats
	cached_stats ContainerMainStats
}

func NewContainerDatum(base types.Container, stats_stream types.ContainerStats) ContainerDatum {
	return ContainerDatum{
		base:         base,
		stats_stream: stats_stream,
		cached_stats: getNewStats(&stats_stream),
	}
}

func (data *ContainerDatum) ID() string {
	return data.base.ID
}

func (data *ContainerDatum) Image() string {
	return data.base.Image
}

func (data *ContainerDatum) UpdatedStats() ContainerMainStats {
	data.cached_stats = getNewStats(&data.stats_stream)
	return data.cached_stats
}

func (data *ContainerDatum) CachedStats() ContainerMainStats {
	return data.cached_stats
}

func (data *ContainerDatum) Close() {
	data.stats_stream.Body.Close()
}

func getNewStats(stats_stream *types.ContainerStats) ContainerMainStats {
	const max_container_stats_data_len = 1 << 14
	container_stats_json_data := make([]byte, max_container_stats_data_len)
	_, err := stats_stream.Body.Read(container_stats_json_data)
	if err != nil && err != io.EOF {
		log.Fatal(err)
	}
	nl_index := utils.FindByte('\n', container_stats_json_data)
	data_json_stats := container_stats_json_data[:nl_index]
	var container_stats_data ContainerMainStats
	if err := json.Unmarshal(data_json_stats, &container_stats_data); err != nil && err != io.EOF {
		log.Fatal(err)
	}
	return container_stats_data
}