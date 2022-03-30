package docker

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/gdamore/tcell/v2"
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

func OpenShellOld(id string, ctx context.Context, shell string) error {
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
		return err
	}
	highjacked_conn, err := docker_cli.ContainerExecAttach(shell_ctx, exec_id.ID, types.ExecStartCheck{Tty: true})
	if err != nil {
		return err
	}
	defer highjacked_conn.Close()
	err = readinessChecker(shell_ctx, exec_id.ID)
	if err != nil {
		return err
	}

	// TODO: replace with log style window
	screen, err := tcell.NewTerminfoScreenFromTty(nil)
	if err != nil {
		log.Fatal("Failed to allocate screen from tty")
	}
	err = screen.Init()
	if err != nil {
		log.Fatal("Failed to init tty screen")
	}
	defer screen.Fini()
	//

	fmt.Printf("Using %s inside container '%s'\n\r", shell, id)

	go screenWriter(&highjacked_conn, shell_ctx)
	go livenessChecker(screen, shell_ctx, exec_id.ID)
	inputParser(screen, highjacked_conn, shell_ctx)

	shell_cancel()

	return err
}

func inputParser(screen tcell.Screen, highjacked_conn types.HijackedResponse, context context.Context) {
	for {
		select {
		case <-context.Done():
			return
		default:
		}
		ev := screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			key := ev.Key()
			switch key {
			case tcell.KeyEnter:
				highjacked_conn.Conn.Write([]byte{0xA})
			case tcell.KeyTab:
				highjacked_conn.Conn.Write([]byte{0x9})
			case tcell.KeyBackspace:
				highjacked_conn.Conn.Write([]byte{0x8}) // does nothing?
			case tcell.KeyBackspace2:
				highjacked_conn.Conn.Write([]byte{0x8})
			case tcell.KeyDelete:
				highjacked_conn.Conn.Write([]byte{27})
				highjacked_conn.Conn.Write([]byte{91})
				highjacked_conn.Conn.Write([]byte{51})
				highjacked_conn.Conn.Write([]byte{126})
			case tcell.KeyCtrlC:
				highjacked_conn.Conn.Write([]byte{0x3})
			case tcell.KeyCtrlD:
				highjacked_conn.Conn.Write([]byte{0x4})
			case tcell.KeyCtrlZ:
				highjacked_conn.Conn.Write([]byte{0x1A})
			case tcell.KeyCtrlR:
				highjacked_conn.Conn.Write([]byte{0x12})
			case tcell.KeyUp:
				highjacked_conn.Conn.Write([]byte{27})
				highjacked_conn.Conn.Write([]byte{79})
				highjacked_conn.Conn.Write([]byte{65})
			case tcell.KeyDown:
				highjacked_conn.Conn.Write([]byte{27})
				highjacked_conn.Conn.Write([]byte{79})
				highjacked_conn.Conn.Write([]byte{66})
			case tcell.KeyLeft:
				highjacked_conn.Conn.Write([]byte{27})
				highjacked_conn.Conn.Write([]byte{79})
				highjacked_conn.Conn.Write([]byte{68})
			case tcell.KeyRight:
				highjacked_conn.Conn.Write([]byte{27})
				highjacked_conn.Conn.Write([]byte{79})
				highjacked_conn.Conn.Write([]byte{67})
			case tcell.KeyPgUp:
				highjacked_conn.Conn.Write([]byte{27})
				highjacked_conn.Conn.Write([]byte{91})
				highjacked_conn.Conn.Write([]byte{53})
				highjacked_conn.Conn.Write([]byte{126})
			case tcell.KeyPgDn:
				highjacked_conn.Conn.Write([]byte{27})
				highjacked_conn.Conn.Write([]byte{91})
				highjacked_conn.Conn.Write([]byte{54})
				highjacked_conn.Conn.Write([]byte{126})
			case tcell.KeyHome:
				highjacked_conn.Conn.Write([]byte{27})
				highjacked_conn.Conn.Write([]byte{91})
				highjacked_conn.Conn.Write([]byte{49})
				highjacked_conn.Conn.Write([]byte{126})
			case tcell.KeyEnd:
				highjacked_conn.Conn.Write([]byte{27})
				highjacked_conn.Conn.Write([]byte{91})
				highjacked_conn.Conn.Write([]byte{52})
				highjacked_conn.Conn.Write([]byte{126})
			case tcell.KeyRune:
				a := string(ev.Rune())
				// if strings.ContainsAny(a, "\n") {
				// 	continue
				// }
				// log.Println(a)
				highjacked_conn.Conn.Write([]byte(a))
			}
		case stopExecEvent:
			log.Println("Stopped exec")
			return
		}
	}
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

func screenWriter(highjacked_conn *types.HijackedResponse, context context.Context) {
	for {
		select {
		case <-context.Done():
			return
		default:
		}
		var buff [1024]byte
		highjacked_conn.Reader.Read(buff[:])
		fmt.Print(string(buff[:]))
	}
}

func livenessChecker(screen tcell.Screen, context context.Context, exec_id string) {
	for {
		select {
		case <-context.Done():
			screen.PostEvent(stopExecEvent{t: time.Now()})
			return
		default:
		}
		exec_inspect, err := docker_cli.ContainerExecInspect(context, exec_id)
		if err != nil {
			log.Fatalf("Failed to inspect exec %s", exec_id)
		}
		if !exec_inspect.Running {
			screen.PostEvent(stopExecEvent{t: time.Now()})
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
}

type stopExecEvent struct {
	t time.Time
}

func (e stopExecEvent) When() time.Time {
	return e.t
}
