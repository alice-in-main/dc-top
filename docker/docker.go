package docker

import (
	"context"
	"io"
	"log"

	"github.com/docker/docker/api/types"
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

func GetContainers() []ContainerData {
	containers, err := docker_cli.ContainerList(context.Background(), types.ContainerListOptions{All: true})
	if err != nil {
		log.Fatal(err)
	}

	num_containers := len(containers)
	container_data := make([]ContainerData, num_containers)

	for i, container := range containers {
		container_id := container.ID
		container_stats, err := docker_cli.ContainerStats(context.Background(), container_id, true)
		if err != nil && err != io.EOF {
			log.Println(container_stats)
			log.Fatal(err)
		}
		container_data[i] = NewContainerData(container, container_stats)
	}

	return container_data
}

func GetDockerInfo() DockerInfo {
	docker_info, err := docker_cli.Info(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	return NewDockerInfo(docker_info)
}
