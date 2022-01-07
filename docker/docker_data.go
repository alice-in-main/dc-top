package docker

import (
	"github.com/docker/docker/api/types"
)

type DockerInfo struct {
	Info types.Info
}

func NewDockerInfo(docker_info types.Info) DockerInfo {
	return DockerInfo{
		Info: docker_info,
	}
}
