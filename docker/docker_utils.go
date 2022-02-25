package docker

import (
	"context"
	"encoding/json"
	"log"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
)

func isBeingRemoved(id string) bool {
	var id_and_status_removing_filter = filters.NewArgs(filters.Arg("id", id), filters.Arg("status", "removing"))
	return isExisting(id, id_and_status_removing_filter)
}

func isDeleted(id string) bool {
	var only_id_filter filters.Args = filters.NewArgs(filters.Arg("id", id))
	return !isExisting(id, only_id_filter)
}

func isExisting(id string, filters filters.Args) bool {
	c, err := docker_cli.ContainerList(context.Background(), types.ContainerListOptions{All: true, Quiet: true, Filters: filters})
	if err != nil {
		log.Fatal(err)
	}
	log.Println(c)
	return len(c) > 0
}

func UsagePercentage(usage int64, limit int64) float64 {
	return 100.0 * float64(usage) / float64(limit)
}

func CpuUsagePercentage(cpu *CpuStats, precpu *CpuStats, inspect_data *types.ContainerJSON) float64 {
	var limit int64
	if inspect_data.HostConfig.NanoCPUs != 0 {
		limit = inspect_data.HostConfig.NanoCPUs
	} else {
		limit = cpu.SystemUsage - precpu.SystemUsage
	}
	return UsagePercentage(cpu.ContainerUsage.TotalUsage-precpu.ContainerUsage.TotalUsage, limit)
}

func MemoryUsagePercentage(mem *MemoryStats) float64 {
	return UsagePercentage(mem.Usage, mem.Limit)
}

func GetEmptyContainerJson() types.ContainerJSON {
	return types.ContainerJSON{ContainerJSONBase: nil}
}

func IsEmptyContainerJson(c *types.ContainerJSON) bool {
	return c.ContainerJSONBase == nil
}

func GetEmptyContainerStats() ContainerMainStats {
	return ContainerMainStats{Name: "!"} // illegal name for real containers
}

func IsEmptyContainerStats(c *ContainerMainStats) bool {
	return c.Name == "!"
}

func NetworkUsageToMapOfInt(s NetworkUsage) map[string]int {
	j, err := json.Marshal(s)
	if err != nil {
		log.Fatalln("Couldn't marshal ", s)
	}
	var new_map map[string]int
	err = json.Unmarshal(j, &new_map)
	if err != nil {
		log.Fatalln("Couldn't unmarshal ", j)
	}
	return new_map
}
