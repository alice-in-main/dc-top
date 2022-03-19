package docker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
)

func isBeingRemoved(ctx context.Context, id string) (bool, error) {
	var id_and_status_removing_filter = filters.NewArgs(filters.Arg("id", id), filters.Arg("status", "removing"))
	return isExisting(ctx, id, id_and_status_removing_filter)
}

func isDeleted(ctx context.Context, id string) (bool, error) {
	var only_id_filter filters.Args = filters.NewArgs(filters.Arg("id", id))
	exists, err := isExisting(ctx, id, only_id_filter)
	return !exists, err
}

func isExisting(ctx context.Context, id string, filters filters.Args) (bool, error) {
	c, err := docker_cli.ContainerList(ctx, types.ContainerListOptions{All: true, Quiet: true, Filters: filters})
	if err != nil {
		return false, err
	}
	return len(c) > 0, nil
}

func UsagePercentage(usage int64, limit int64) float64 {
	return 100.0 * float64(usage) / float64(limit)
}

func CpuUsagePercentage(cpu *CpuStats, precpu *CpuStats, inspect_data *types.ContainerJSON) float64 {
	var limit int64
	if inspect_data.ContainerJSONBase != nil && inspect_data.HostConfig.NanoCPUs != 0 {
		limit = inspect_data.HostConfig.NanoCPUs
	} else {
		limit = cpu.SystemUsage - precpu.SystemUsage
	}
	return UsagePercentage(cpu.ContainerUsage.TotalUsage-precpu.ContainerUsage.TotalUsage, limit)
}

func MemoryUsagePercentage(mem *MemoryStats) float64 {
	return UsagePercentage(mem.Usage, mem.Limit)
}

func NetworkUsageToMapOfInt(s NetworkUsage) (map[string]int, error) {
	j, err := json.Marshal(s)
	if err != nil {
		return nil, fmt.Errorf("couldn't marshal %+v", s)
	}
	var new_map map[string]int
	err = json.Unmarshal(j, &new_map)
	if err != nil {
		return nil, fmt.Errorf("couldn't unmarshal %s", j)
	}
	return new_map, nil
}
