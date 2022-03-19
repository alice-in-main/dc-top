package subshells

import (
	"context"
	"dc-top/docker/compose"
	"dc-top/gui/view/window"
	"dc-top/gui/view/window/bar_window"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/creack/pty"
)

func OpenContainerShell(id string, ctx context.Context) {
	screen := window.GetScreen()
	screen.PostEvent(window.NewPauseWindowsEvent())
	go func() {
		cmd := exec.CommandContext(ctx, "docker", "exec", "-it", id, "sh")
		err := runCmdInPty(cmd, "clear\n")
		if err != nil {
			log.Fatal(err)
		}
		screen.PostEvent(window.NewResumeWindowsEvent())
	}()
}

func EditDcYaml(ctx context.Context) {
	screen := window.GetScreen()
	if compose.DcModeEnabled() {
		screen.PostEvent(window.NewPauseWindowsEvent())
		go func() {
			compose.CreateBackupYaml()
			cmd := exec.CommandContext(ctx, "vim", compose.DcYamlPath())
			err := runCmdInPty(cmd)
			if err != nil {
				log.Fatal(err)
			}
			screen.PostEvent(window.NewResumeWindowsEvent())
			if compose.ValidateYaml(ctx) {
				bar_window.Info([]rune("restarting docker-compose"))
				compose.Up(ctx)
			} else {
				bar_window.Info([]rune("docker compose yaml is invalid"))
				compose.RestoreFromBackup()
			}
		}()
	} else {
		bar_window.Info([]rune("docker compose mode is disabled"))
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
