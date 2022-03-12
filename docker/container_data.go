package docker

import (
	"context"
	"dc-top/docker/compose"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"sort"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
)

type ContainerData struct {
	data                []ContainerDatum
	main_sort_type      SortType
	secondary_sort_type SortType
	// only with docker-compose mode
	dc_services *compose.Services
	dc_filters  *filters.Args
}

type SortType uint8

const (
	Name SortType = iota
	Memory
	Cpu
	Image
	State
	None
)

func NewContainerData() (ContainerData, error) {
	var container_list_options = types.ContainerListOptions{All: true}
	var dc_services *compose.Services = nil
	var dc_filters *filters.Args = nil
	var err error
	if compose.DcModeEnabled() {
		dc_services = &compose.Services{}
		*dc_services, err = compose.GenerateDcData()
		if err != nil {
			return ContainerData{}, err
		}
		dc_filters = &filters.Args{}
		*dc_filters = filters.NewArgs()
		for service_key, service := range dc_services.ServicesMap {
			if service.ContainerName == "" {
				dc_filters.Add("name", fmt.Sprintf("%s_%s_1", filepath.Base(filepath.Dir(compose.DcYamlPath())), service_key))
			} else {
				dc_filters.Add("name", service.ContainerName)
			}
		}
		container_list_options.Filters = *dc_filters
	}
	containers, err := docker_cli.ContainerList(context.TODO(), container_list_options)
	if err != nil {
		return ContainerData{}, err
	}

	container_data := make([]ContainerDatum, 0)
	container_init_ch := make(chan error, len(containers))
	defer close(container_init_ch)

	for index, container := range containers {
		go func(i int, c types.Container) {
			container_id := c.ID
			container_stats, err := docker_cli.ContainerStats(context.Background(), container_id, true)
			if err != nil && err != io.EOF {
				log.Println(err)
				if !strings.HasPrefix(err.Error(), "Error response from daemon: No such container") {
					log.Println(containers)
				}
			} else {
				new_datum, _err := NewContainerDatum(c, container_stats)
				container_data = append(container_data, new_datum)
				err = _err
			}
			container_init_ch <- err
		}(index, container)
	}
	for range containers {
		_err := <-container_init_ch
		if err != nil {
			err = _err
		}
	}

	new_containers_data := ContainerData{
		data:                container_data,
		main_sort_type:      State,
		secondary_sort_type: Name,
		dc_services:         dc_services,
		dc_filters:          dc_filters,
	}

	return new_containers_data, err
}

// func (containers *ContainerData) Clone() ContainerData {
// 	// var copy ContainerData
// 	// copier.CopyWithOption(&copy, containers, copier.Option{IgnoreEmpty: true, DeepCopy: true})
// 	// return copy
// 	clone := *containers
// 	copy(clone.data, containers.data)

// 	return clone
// }

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

func (containers *ContainerData) GetSortedData(main_sort_type, secondary_sort_type SortType) ContainerData {
	var data_copy = make([]ContainerDatum, containers.Len())
	copy(data_copy, containers.data)

	new_data := ContainerData{
		data:                data_copy,
		main_sort_type:      main_sort_type,
		secondary_sort_type: secondary_sort_type,
		dc_services:         containers.dc_services,
		dc_filters:          containers.dc_filters,
	}
	sort.Stable(&new_data)

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

func (containers *ContainerData) UpdateStats() {
	data := containers.GetData()
	for i, datum := range data {
		new_datum, err := UpdatedDatum(datum)
		if err != nil {
			log.Printf("Got error %s while fetching new data", err)
		}

		containers.data[i] = new_datum
	}
}

func (containers *ContainerData) GetUpdatedStats() ContainerData {
	new_data := *containers
	new_data.data = make([]ContainerDatum, containers.Len())
	copy(new_data.data, containers.data)
	for i, datum := range containers.data {
		new_datum, err := UpdatedDatum(datum)
		if err != nil {
			log.Printf("Got error %s while fetching new data", err)
		}

		new_data.data[i] = new_datum
	}
	return new_data
}

func (containers *ContainerData) AreIdsUpToDate() (bool, error) {
	preserved_container_options := types.ContainerListOptions{All: true, Quiet: true, Filters: filters.NewArgs()}
	all_containers_options := types.ContainerListOptions{All: true, Quiet: true}

	if containers.dc_filters != nil {
		preserved_container_options.Filters = containers.dc_filters.Clone()
		all_containers_options.Filters = containers.dc_filters.Clone()
	}

	for _, c := range containers.data {
		preserved_container_options.Filters.Add("id", c.base.ID)
	}

	var err error
	preserved_containers, err := docker_cli.ContainerList(context.Background(), preserved_container_options)
	if err != nil {
		return false, err
	}
	all_containers, err := docker_cli.ContainerList(context.Background(), all_containers_options)
	if err != nil {
		return false, err
	}

	return len(preserved_containers) == containers.Len() && len(all_containers) == containers.Len(), nil
}

func (containers *ContainerData) Contains(id string) bool {
	for _, c := range containers.data {
		if c.base.ID == id {
			return true
		}
	}
	return false
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
			return usage_i > usage_j
		}
	case Cpu:
		{
			stats_i := i.CachedStats()
			stats_j := j.CachedStats()
			usage_i := CpuUsagePercentage(&stats_i.Cpu, &stats_i.PreCpu, &i.inspection)
			usage_j := CpuUsagePercentage(&stats_j.Cpu, &stats_j.PreCpu, &j.inspection)
			return usage_i > usage_j
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
