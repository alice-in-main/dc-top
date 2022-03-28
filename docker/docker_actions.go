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
	MaxSavedLogs = 1000
)

func PauseContainer(ctx context.Context, id string) error {
	return docker_cli.ContainerPause(ctx, id)
}

func UnpauseContainer(ctx context.Context, id string) error {
	return docker_cli.ContainerUnpause(ctx, id)
}

func StopContainer(ctx context.Context, id string) error {
	duration := 3 * time.Second
	return docker_cli.ContainerStop(ctx, id, &duration)
}

func RestartContainer(ctx context.Context, id string) error {
	duration := 10 * time.Second
	return docker_cli.ContainerRestart(ctx, id, &duration)
}

func DeleteContainer(ctx context.Context, id string) error {
	return docker_cli.ContainerRemove(ctx, id,
		types.ContainerRemoveOptions{RemoveVolumes: true, RemoveLinks: false, Force: true})
}

func StreamContainerLogs(id string, writer io.Writer, ctx context.Context, cancel context.CancelFunc) {
	reader, err := docker_cli.ContainerLogs(ctx, id, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Tail:       fmt.Sprintf("%d", MaxSavedLogs),
		Follow:     true,
	})
	if err != nil {
		log.Println(err)
		cancel()
		return
	}
	defer reader.Close()

	_, err = io.Copy(writer, reader)
	if err != nil && err != io.EOF && err.Error() != "context canceled" {
		log.Println(err)
		cancel()
	}
}

func InspectContainerNoPanic(ctx context.Context, id string) types.ContainerJSON {
	j, err := docker_cli.ContainerInspect(ctx, id)
	if err != nil {
		return types.ContainerJSON{}
	}
	return j
}
