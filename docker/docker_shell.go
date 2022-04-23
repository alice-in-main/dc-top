package docker

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/docker/docker/api/types"
)

func OpenShell(id string, ctx context.Context, shell string) (*types.HijackedResponse, error) {
	var cfg = types.ExecConfig{
		Tty:          true,
		AttachStdin:  true,
		AttachStderr: true,
		AttachStdout: true,
		Cmd:          []string{shell},
	}
	shell_ctx, shell_cancel := context.WithCancel(ctx)
	defer shell_cancel()

	exec_id, err := docker_cli.ContainerExecCreate(shell_ctx, id, cfg)
	if err != nil {
		return nil, err
	}
	highjacked_conn, err := docker_cli.ContainerExecAttach(shell_ctx, exec_id.ID, types.ExecStartCheck{Tty: true})
	if err != nil {
		return nil, err
	}
	err = readinessChecker(shell_ctx, exec_id.ID)
	if err != nil {
		return nil, err
	}

	log.Printf("Using %s inside container '%s'\n\r", shell, id)
	return &highjacked_conn, nil
}

func readinessChecker(context context.Context, exec_id string) error {
	for {
		exec_inspect, err := docker_cli.ContainerExecInspect(context, exec_id)
		if err != nil {
			return fmt.Errorf("failed to inspect exec %s", exec_id)
		}
		if exec_inspect.ExitCode != 0 || !exec_inspect.Running {
			return errors.New("invalid entrypoint")
		}
		if exec_inspect.Pid == 0 {
			time.Sleep(5 * time.Millisecond)
			continue
		}
		break
	}
	return nil
}
