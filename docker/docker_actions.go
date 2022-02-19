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

func StreamContainerLogs(id string, writer io.Writer, c context.Context) error {
	reader, err := docker_cli.ContainerLogs(c, id, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Tail:       fmt.Sprintf("%d", NumSavedLogs),
		Follow:     true,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()

	_, err = io.Copy(writer, reader)
	if err != nil && err != io.EOF && err.Error() != "context canceled" {
		log.Fatal(err)
	}

	return nil
}

func InspectContainer(id string) types.ContainerJSON {
	j, err := docker_cli.ContainerInspect(context.Background(), id)
	if err != nil {
		log.Fatalf("Got error '%s' when inspecting '%s'", err, id)
	}
	return j
}