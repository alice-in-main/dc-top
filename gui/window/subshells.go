package window

import (
	"context"
	"dc-top/docker/compose"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/creack/pty"
	"github.com/gdamore/tcell/v2"
)

func OpenContainerShell(id string, ctx context.Context, screen tcell.Screen) {
	screen.PostEvent(NewPauseWindowsEvent())
	go func() {
		cmd := exec.CommandContext(context.TODO(), "docker", "exec", "-it", id, "sh")
		err := runCmdInPty(cmd, "clear\n")
		if err != nil {
			log.Fatal(err)
		}
		screen.PostEvent(NewResumeWindowsEvent())
	}()
}

func EditDcYaml(ctx context.Context, screen tcell.Screen) {
	if compose.DcModeEnabled() {
		screen.PostEvent(NewPauseWindowsEvent())
		go func() {
			compose.CreateBackupYaml()
			cmd := exec.CommandContext(context.TODO(), "vim", compose.DcYamlPath())
			err := runCmdInPty(cmd)
			if err != nil {
				log.Fatal(err)
			}
			screen.PostEvent(NewResumeWindowsEvent())
			if compose.ValidateYaml(context.TODO()) {
				screen.PostEvent(NewMessageEvent(Bar, ContainersHolder, InfoMessage{Msg: []rune("restarting docker-compose")}))
				compose.Up(context.TODO())
			} else {
				screen.PostEvent(NewMessageEvent(Bar, ContainersHolder, ErrorMessage{Msg: []rune("docker compose yaml is invalid")}))
				compose.RestoreFromBackup()
			}
		}()
	} else {
		screen.PostEvent(NewMessageEvent(Bar, ContainersHolder, ErrorMessage{Msg: []rune("docker compose mode is disabled")}))
	}
}

func runCmdInPty(command *exec.Cmd, args ...string) error {
	// Start the command with a pty.
	ptmx, err := pty.Start(command)
	if err != nil {
		return err
	}
	// Make sure to close the pty at the end.
	defer func() { _ = ptmx.Close() }() // Best effort.

	// Handle pty size.
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)
	go func() {
		for range ch {
			if err := pty.InheritSize(os.Stdin, ptmx); err != nil {
				log.Printf("error resizing pty: %s", err)
			}
		}
	}()
	ch <- syscall.SIGWINCH                        // Initial resize.
	defer func() { signal.Stop(ch); close(ch) }() // Cleanup signals when done.

	// Set stdin in raw mode.
	// old_state, err := term.MakeRaw(int(os.Stdin.Fd()))
	// if err != nil {
	// 	return err
	// }
	// defer func() { _ = term.Restore(int(os.Stdin.Fd()), old_state) }() // Best effort.

	for _, arg := range args {
		io.WriteString(ptmx, arg)
	}
	// Copy stdin to the pty and the pty to stdout.
	// NOTE: The goroutine will keep reading until the next keystroke before returning.
	go func() { _, _ = io.Copy(ptmx, os.Stdin) }()
	_, _ = io.Copy(os.Stdout, ptmx)

	return nil
}
