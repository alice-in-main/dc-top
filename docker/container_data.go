package docker

import (
	"dc-top/utils"
	"encoding/json"
	"fmt"
	"io"
	"log"

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

const max_container_stats_data_len = 1 << 13

func (data ContainerData) String() string {
	container_stats_json_data := make([]byte, max_container_stats_data_len)
	container_id := data.base.ID[:10]
	_, err := data.stats.Body.Read(container_stats_json_data)
	if err != nil && err != io.EOF {
		log.Fatal(err)
	}
	nl_index := utils.FindByte('\n', container_stats_json_data)
	relevant_container_stats_json_data := container_stats_json_data[:nl_index]
	var container_stats_data map[string]interface{}
	if err := json.Unmarshal(relevant_container_stats_json_data, &container_stats_data); err != nil && err != io.EOF {
		log.Println(string(relevant_container_stats_json_data))
		log.Fatal(err)
	}
	return fmt.Sprintf("%s %s %s %s %s", container_id, data.base.Image, data.stats.OSType, data.base.Status, container_stats_data["read"])
}

func (data ContainerData) Close() {
	data.stats.Body.Close()
}
