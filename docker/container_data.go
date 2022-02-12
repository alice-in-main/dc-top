package docker

import (
	"context"
	"io"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
)

type ContainerData struct {
	data                []ContainerDatum
	main_sort_type      SortType
	secondary_sort_type SortType
}

type SortType int8

const (
	Name SortType = iota
	Memory
	Cpu
	Image
	State
	None
)

func NewContainerData() ContainerData {
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
			for err != nil && err != io.EOF {
				log.Println(err)
				if !strings.HasPrefix(err.Error(), "Error response from daemon: No such container") {
					log.Println(containers)
					panic(1)
				}
				container_stats, err = docker_cli.ContainerStats(context.Background(), container_id, true)
			}
			container_data[i] = NewContainerDatum(c, container_stats)
			container_init_ch <- i
		}(index, container)
	}
	for range container_data {
		<-container_init_ch
	}

	new_containers_data := ContainerData{
		data:                container_data,
		main_sort_type:      State,
		secondary_sort_type: Name,
	}

	return new_containers_data
}

func (containers *ContainerData) Len() int {
	return len(containers.data)
}

func (containers *ContainerData) Less(i, j int) bool {
	if containers.GetData()[i].ID() == containers.GetData()[j].ID() {
		log.Fatal("Shouldn't get here")
	}
	// return j < i
	// return containers.GetData()[i].ID() < containers.GetData()[j].ID()
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

func assertNoDuplicates(containers_data ContainerData, message string) {
	for i, c1 := range containers_data.GetData()[:containers_data.Len()-1] {
		for _, c2 := range containers_data.GetData()[i+1:] {
			if c1.ID() == c2.ID() {
				log.Printf("%s: Found duplicate ids", message)
				//log.Fatal(c1, c2)
			}
		}
	}
}

func (containers *ContainerData) SortData(main_sort_type, secondary_sort_type SortType) {
	start := time.Now()

	containers.main_sort_type = main_sort_type
	containers.secondary_sort_type = secondary_sort_type
	sort.Stable(containers)

	assertNoDuplicates(*containers, "inside sort data")

	elapsed := time.Since(start)
	log.Printf("It took %dmicrosecconds to sort data", elapsed.Microseconds())
}

func (containers *ContainerData) UpdateStats() {
	var set_of_ids map[string]ContainerDatum = make(map[string]ContainerDatum)
	assertNoDuplicates(*containers, "Inside update stats, before loop")
	data := containers.GetData()
	for i, datum := range data {
		if d, ok := set_of_ids[datum.ID()]; ok {
			log.Println(d, datum)
			//log.Fatalf("%s already exists", datum.ID())
		} else {
			log.Printf("Didnt find %s", datum.ID())
		}
		new_datum, err := UpdatedDatum(datum)
		if err != nil {
			log.Printf("Got error %s while fetching new data", err)
		}
		set_of_ids[datum.ID()] = datum

		containers.data[i] = new_datum
	}
	assertNoDuplicates(*containers, "Inside update stats, after loop")
}

func (containers *ContainerData) AreIdsUpToDate() bool {
	var filtered_ids filters.Args = filters.NewArgs()
	for _, c := range containers.data {
		filtered_ids.Add("id", c.base.ID)
	}

	var err error
	preserved_containers, err := docker_cli.ContainerList(context.Background(), types.ContainerListOptions{All: true, Quiet: true, Filters: filtered_ids})
	if err != nil {
		log.Fatal(err)
	}
	all_containers, err := docker_cli.ContainerList(context.Background(), types.ContainerListOptions{All: true, Quiet: true})
	if err != nil {
		log.Fatal(err)
	}

	return len(preserved_containers) == containers.Len() && len(all_containers) == containers.Len()
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
			return usage_i < usage_j
		}
	case Cpu:
		{
			stats_i := i.CachedStats()
			stats_j := j.CachedStats()
			usage_i := CpuUsagePercentage(&stats_i.Cpu, &stats_i.PreCpu)
			usage_j := CpuUsagePercentage(&stats_j.Cpu, &stats_j.PreCpu)
			return usage_i < usage_j
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
