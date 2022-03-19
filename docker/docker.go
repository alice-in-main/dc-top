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

func GetContainers(old_data *ContainerData) (ContainerData, error) {
	if old_data != nil {
		for _, datum := range old_data.data {
			datum.Close()
		}
	}

	return NewContainerData()
}

func GetDockerInfo() (DockerInfo, error) {
	docker_info, err := docker_cli.Info(context.Background())
	if err != nil {
		return DockerInfo{}, err
	}
	return NewDockerInfo(docker_info), nil
}

func Close() {
	docker_cli.Close()
}
