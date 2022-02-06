package docker

import (
	"context"
	"log"

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

func GetContainers(old_data *ContainerData) ContainerData {
	if old_data != nil {
		for _, datum := range old_data.data {
			datum.Close()
		}
	}

	return NewContainerData(SortType(Name))
}

func GetDockerInfo() DockerInfo {
	docker_info, err := docker_cli.Info(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	return NewDockerInfo(docker_info)
}
