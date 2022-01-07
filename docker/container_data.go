package docker

import (
	"bytes"
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

var max_container_stats_data_len = 16384
var container_stats_json_data []byte = make([]byte, max_container_stats_data_len)

func (data ContainerData) String() string {
	container_stats_json_data = bytes.Repeat([]byte("\x00"), max_container_stats_data_len)
	container_id := data.base.ID[:10]
	n, err := data.stats.Body.Read(container_stats_json_data)
	if err != nil && err != io.EOF {
		panic(err)
	}
	container_stats_json_data = container_stats_json_data[:n]
	var container_stats_data map[string]interface{}
	if err := json.Unmarshal(container_stats_json_data, &container_stats_data); err != nil && err != io.EOF {
		panic(err)
	}
	return fmt.Sprintf("%s %s %s %s", container_id, data.base.Image, data.stats.OSType, container_stats_data["read"])
}
