package docker

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/docker/docker/api/types"
)

const (
	NumSavedLogs = 1000
)

func DeleteContainer(id string) error {
	return docker_cli.ContainerRemove(context.Background(), id,
		types.ContainerRemoveOptions{RemoveVolumes: true, RemoveLinks: false, Force: true})
}

func StreamContainerLogs(id string, writer io.Writer, window context.Context) error {
	reader, err := docker_cli.ContainerLogs(context.Background(), id, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Tail:       fmt.Sprintf("%d", NumSavedLogs),
		Follow:     true,
	})
	if err != nil {
		log.Fatal(err)
	}

	_, err = io.Copy(writer, reader)
	if err != nil && err != io.EOF {
		log.Fatal(err)
	}
	reader.Close()
	return nil
}
