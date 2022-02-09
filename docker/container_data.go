package docker

import (
	"context"
	"io"
	"log"
	"sort"
	"strings"

	"github.com/docker/docker/api/types"
)

type ContainerData struct {
	data      []ContainerDatum
	sorted_by SortType
}

type SortType int8

const (
	Name SortType = iota
	Memory
	Cpu
	Image
	State
)

func NewContainerData(sort_type SortType) ContainerData {
	containers, err := docker_cli.ContainerList(context.Background(), types.ContainerListOptions{All: true})
	if err != nil {
		log.Fatal(err)
	}

	num_containers := len(containers)
	container_data := make([]ContainerDatum, num_containers)
	container_init_ch := make(chan interface{}, num_containers)
	defer close(container_init_ch)

	for index, container := range containers {
		go func(i int, c types.Container) {
			container_id := c.ID
			container_stats, err := docker_cli.ContainerStats(context.Background(), container_id, true)
			if err != nil && err != io.EOF {
				log.Println(container_stats)
				log.Fatal(err)
			}
			container_data[i] = NewContainerDatum(c, container_stats)
			container_init_ch <- i
		}(index, container)
	}
	for range container_data {
		<-container_init_ch
	}

	new_containers_data := ContainerData{
		data:      container_data,
		sorted_by: sort_type,
	}
	new_containers_data.SortData(sort_type)

	return new_containers_data
}

func (containers *ContainerData) Len() int {
	return len(containers.data)
}

func (containers *ContainerData) Less(i, j int) bool {
	switch containers.sorted_by {
	case Name:
		{
			name_i := containers.data[i].CachedStats().Name
			name_j := containers.data[j].CachedStats().Name
			return strings.Compare(name_i, name_j) == -1
		}
	case Memory:
		{
			memory_i := containers.data[i].CachedStats().Memory
			usage_i := float64(memory_i.Usage) / float64(memory_i.Limit)
			memory_j := containers.data[j].CachedStats().Memory
			usage_j := float64(memory_j.Usage) / float64(memory_j.Limit)
			return usage_i < usage_j
		}
	case State:
		{
			return containers.data[i].base.State == "running" && containers.data[j].base.State != "running"
		}
	default:
		log.Println("Unimplemented sort type")
		panic(1)
	}
}

func (containers *ContainerData) Swap(i, j int) {
	containers.data[i], containers.data[j] = containers.data[j], containers.data[i]
}

func (containers *ContainerData) GetData() []ContainerDatum {
	return containers.data
}

func (containers *ContainerData) SortData(sort_type SortType) {
	containers.sorted_by = sort_type
	sort.Stable(containers)
}
