package docker

import (
	"context"
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

// func isBeingCreated(id string) bool {
// 	var id_and_status_created_filter = filters.NewArgs(filters.Arg("id", id), filters.Arg("status", "created"))
// 	return isExisting(id, id_and_status_created_filter)
// }

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

func CpuUsagePercentage(cpu *CpuStats, precpu *CpuStats) float64 {
	return UsagePercentage(cpu.ContainerUsage.TotalUsage-precpu.ContainerUsage.TotalUsage, cpu.SystemUsage-precpu.SystemUsage)
}

func MemoryUsagePercentage(mem *MemoryStats) float64 {
	return UsagePercentage(mem.Usage, mem.Limit)
}
