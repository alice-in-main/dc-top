package docker

import (
	"context"
	"dc-top/docker/compose"
	"fmt"
	"io"
	"log"
	"math"
	"sort"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
)

type ContainerData struct {
	data                []ContainerDatum
	main_sort_type      SortType
	secondary_sort_type SortType
}

func NewContainerData(ctx context.Context) (ContainerData, error) {
	containers_options := types.ContainerListOptions{All: true, Quiet: true}
	filters, err := getContainerFilters(ctx)
	if err != nil {
		log.Println("Failed to generate filter data when getting new data")
		return ContainerData{}, err
	}
	containers_options.Filters = filters

	containers, err := docker_cli.ContainerList(ctx, containers_options)
	if err != nil {
		log.Println("Failed to get container list on getting new data")
		log.Println(containers)
		return ContainerData{}, err
	}

	var new_data = make([]ContainerDatum, 0)
	var data_channel = make(chan *ContainerDatum, len(containers))
	go func() {
		for _, container := range containers {
			go func(_inner_cont types.Container) {
				container_stats, err := docker_cli.ContainerStats(ctx, _inner_cont.ID, true)
				if err != nil && err != io.EOF {
					log.Println(err)
					if !strings.HasPrefix(err.Error(), "Error response from daemon: No such container") {
						log.Printf("%s: %s", err, _inner_cont.ID)
					}
					// data_channel <- nil
				} else {
					new_datum, err := NewContainerDatum(ctx, _inner_cont, container_stats)
					if err != nil {
						new_datum.is_deleted = true
					}
					data_channel <- &new_datum
				}
			}(container)
		}
	}()
	for range containers {
		new_datum := <-data_channel
		if new_datum != nil {
			new_data = append(new_data, *new_datum)
		}
	}

	new_containers_data := ContainerData{
		data:                new_data,
		main_sort_type:      State,
		secondary_sort_type: Name,
	}

	return new_containers_data, err
}

func UpdatedContainerData(ctx context.Context, old_data *ContainerData) (ContainerData, error) {
	containers_options := types.ContainerListOptions{All: true, Quiet: true}
	filters, err := getContainerFilters(ctx)
	if err != nil {
		log.Println("Failed to generate filter data when getting updated data")
		return ContainerData{}, err
	}
	containers_options.Filters = filters

	containers, err := docker_cli.ContainerList(ctx, containers_options)
	if err != nil {
		log.Println("Failed to get all containers data on updating data")
		log.Println(containers)
		return ContainerData{}, err
	}

	var new_data = make([]ContainerDatum, 0)
	var data_channel = make(chan *ContainerDatum, len(containers))

	go func() {
		for _, old_datum := range old_data.GetData() {
			go func(_inner_old_datum ContainerDatum) {
				if base := findContainerBase(&_inner_old_datum, containers); base != nil {
					updated_datum, err := UpdatedDatum(ctx, _inner_old_datum, *base)
					if err != nil {
						updated_datum.is_deleted = true
					}
					data_channel <- &updated_datum
				} else {
					_inner_old_datum.Close()
				}
			}(old_datum)
		}
	}()

	go func() {
		for _, container := range containers {
			go func(_inner_cont types.Container) {
				if !isContainerExists(&_inner_cont, old_data.GetData()) {
					log.Printf("%s doesn't exist", _inner_cont.Image)
					container_stats, err := docker_cli.ContainerStats(ctx, _inner_cont.ID, true)
					if err != nil && err != io.EOF {
						log.Println(err)
						if !strings.HasPrefix(err.Error(), "Error response from daemon: No such container") {
							log.Printf("%s: %s", err, _inner_cont.ID)
						}
						// data_channel <- nil
					} else {
						new_datum, err := NewContainerDatum(ctx, _inner_cont, container_stats)
						if err != nil {
							new_datum.is_deleted = true
						}
						data_channel <- &new_datum
					}
				}
			}(container)
		}
	}()

	for range containers {
		new_datum := <-data_channel
		if new_datum != nil {
			new_data = append(new_data, *new_datum)
		}
	}

	new_containers_data := ContainerData{
		data:                new_data,
		main_sort_type:      State,
		secondary_sort_type: Name,
	}

	return new_containers_data, err
}

func findContainerBase(datum *ContainerDatum, containers []types.Container) *types.Container {
	for _, container := range containers {
		if datum.ID() == container.ID {
			return &container
		}
	}
	return nil
}

func isContainerExists(container *types.Container, data []ContainerDatum) bool {
	for _, datum := range data {
		if datum.ID() == container.ID {
			return true
		}
	}
	return false
}

func (containers *ContainerData) Len() int {
	return len(containers.data)
}

func (containers *ContainerData) Less(i, j int) bool {
	if containers.GetData()[i].ID() == containers.GetData()[j].ID() {
		log.Fatal("Shouldn't get here 3")
	}
	if lessAux(containers.main_sort_type, &containers.GetData()[i], &containers.GetData()[j]) {
		return true
	}
	if lessAux(containers.main_sort_type, &containers.GetData()[j], &containers.GetData()[i]) {
		return false
	}
	return lessAux(containers.secondary_sort_type, &containers.GetData()[i], &containers.GetData()[j])
}

func (containers *ContainerData) Swap(i, j int) {
	containers.data[i], containers.data[j] = containers.data[j], containers.data[i]
}

func (containers *ContainerData) GetData() []ContainerDatum {
	return containers.data
}

func (containers *ContainerData) GetSortedData(main_sort_type, secondary_sort_type SortType, reverse bool) ContainerData {
	var data_copy = make([]ContainerDatum, containers.Len())
	copy(data_copy, containers.data)

	new_data := ContainerData{
		data:                data_copy,
		main_sort_type:      main_sort_type,
		secondary_sort_type: secondary_sort_type,
	}
	if reverse {
		sort.Stable(sort.Reverse(&new_data))
	} else {
		sort.Stable(&new_data)
	}

	return new_data
}

func (containers *ContainerData) Filter(substr string) []ContainerDatum {
	filtered_data := make([]ContainerDatum, 0)
	for _, datum := range containers.GetData() {
		if datum.Contains(substr) {
			filtered_data = append(filtered_data, datum)
		}
	}
	return filtered_data
}

func (containers *ContainerData) Contains(id string) bool {
	for _, c := range containers.data {
		if c.base.ID == id {
			return true
		}
	}
	return false
}

func getContainerFilters(ctx context.Context) (filters.Args, error) {
	var contaiener_filters filters.Args = filters.NewArgs()
	if compose.DcModeEnabled() {
		dc_filters, err := compose.GetDcProcesses(ctx)
		if err != nil {
			return contaiener_filters, err
		}

		for _, filter := range dc_filters {
			contaiener_filters.Add("name", fmt.Sprintf("^/%s$", filter.Name))
		}
	}

	return contaiener_filters, nil
}

var docker_state_priority = map[string]uint8{
	"running":    0,
	"created":    1,
	"restarting": 2,
	"paused":     3,
	"dead":       4,
	"exited":     5,
}

func lessAux(sort_by SortType, i, j *ContainerDatum) bool {
	switch sort_by {
	case Name:
		{
			stats_i := i.CachedStats()
			stats_j := j.CachedStats()
			name_i := stats_i.Name
			name_j := stats_j.Name
			return name_i < name_j
		}
	case Image:
		{
			image_i := i.Image()
			image_j := j.Image()
			return image_i < image_j
		}
	case Memory:
		{
			stats_i := i.CachedStats()
			stats_j := j.CachedStats()
			usage_i := MemoryUsagePercentage(&stats_i.Memory)
			usage_j := MemoryUsagePercentage(&stats_j.Memory)
			return usage_i > usage_j || (!math.IsNaN(usage_i) && math.IsNaN(usage_j))
		}
	case Cpu:
		{
			stats_i := i.CachedStats()
			stats_j := j.CachedStats()
			usage_i := CpuUsagePercentage(&stats_i.Cpu, &stats_i.PreCpu, &i.inspection)
			usage_j := CpuUsagePercentage(&stats_j.Cpu, &stats_j.PreCpu, &j.inspection)
			return usage_i > usage_j || (!math.IsNaN(usage_i) && math.IsNaN(usage_j))
		}
	case State:
		{
			return docker_state_priority[i.base.State] < docker_state_priority[j.base.State]
		}
	default:
		log.Println("Unimplemented sort type")
		panic(1)
	}
}
