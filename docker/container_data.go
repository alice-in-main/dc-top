package docker

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/docker/docker/api/types"
)

type ContainerData struct {
	base  types.Container
	stats types.ContainerStats
}

func NewContainerData(base types.Container, stats types.ContainerStats) ContainerData {
	return ContainerData{
		base:  base,
		stats: stats,
	}
}

var max_container_stats_data_len = 1 << 13
var container_stats_json_data []byte = make([]byte, max_container_stats_data_len)

func (data ContainerData) String() string {
	for i := range container_stats_json_data {
		container_stats_json_data[i] = '\x00'
	}
	container_id := data.base.ID[:10]
	n, err := data.stats.Body.Read(container_stats_json_data)
	if err != nil && err != io.EOF {
		panic(err)
	}
	relevant_container_stats_json_data := container_stats_json_data[:n]
	var container_stats_data map[string]interface{}
	if err := json.Unmarshal(relevant_container_stats_json_data, &container_stats_data); err != nil && err != io.EOF {
		panic(err)
	}
	return fmt.Sprintf("%s %s %s %s %s", container_id, data.base.Image, data.stats.OSType, data.base.Status, container_stats_data["read"])
}

func (data ContainerData) Close() {
	data.stats.Body.Close()
}
