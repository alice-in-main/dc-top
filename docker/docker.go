package docker

import (
	"context"
	"log"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

var docker_cli *client.Client

func Init() {
	var err error
	docker_cli, err = client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		log.Fatal(err)
	}
}

func Test() {
	var err error
	var filtered_ids filters.Args = filters.NewArgs()
	containers, err := docker_cli.ContainerList(context.Background(), types.ContainerListOptions{All: true, Size: true})
	if err != nil {
		log.Fatal(err)
	}
	for _, c := range containers {
		filtered_ids.Add("id", c.ID)
	}

	updated_num_containers, err := docker_cli.ContainerList(context.Background(), types.ContainerListOptions{All: true, Size: true})
	if err != nil {
		log.Fatal(err)
	}
	new_containers, err := docker_cli.ContainerList(context.Background(), types.ContainerListOptions{All: true, Filters: filtered_ids})
	if err != nil {
		log.Fatal(err)
	}

	log.Println(updated_num_containers)
	log.Println(new_containers)
}

func GetContainers(old_data *ContainerData) ContainerData {
	if old_data != nil {
		for _, datum := range old_data.data {
			datum.Close()
		}
	}

	return NewContainerData()
}

func GetDockerInfo() DockerInfo {
	docker_info, err := docker_cli.Info(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	return NewDockerInfo(docker_info)
}
