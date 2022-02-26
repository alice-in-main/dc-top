package docker

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/docker/docker/api/types"
)

const (
	NumSavedLogs = 1000
)

func PauseContainer(id string) error {
	return docker_cli.ContainerPause(context.TODO(), id)
}

func UnpauseContainer(id string) error {
	return docker_cli.ContainerUnpause(context.TODO(), id)
}

func StopContainer(id string) error {
	duration := 3 * time.Second
	return docker_cli.ContainerStop(context.TODO(), id, &duration)
}

func DeleteContainer(id string) error {
	return docker_cli.ContainerRemove(context.TODO(), id,
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

func InspectContainerNoPanic(id string) types.ContainerJSON {
	j, err := docker_cli.ContainerInspect(context.Background(), id)
	if err != nil {
		return types.ContainerJSON{}
	}
	return j
}

func InspectContainer(id string) types.ContainerJSON {
	j, err := docker_cli.ContainerInspect(context.Background(), id)
	if err != nil {
		log.Fatalf("Got error '%s' when inspecting '%s'", err, id)
	}
	return j
}
