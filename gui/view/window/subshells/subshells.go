package subshells

import (
	"context"
	"dc-top/gui/view/window"
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
		err := runCmdInPty(ctx, cmd, "clear\n")
		if err != nil {
			log.Fatal(err)
		}
		screen.PostEvent(window.NewResumeWindowsEvent())
	}()
}

func runCmdInPty(ctx context.Context, command *exec.Cmd, args ...string) error {
	// Start the command with a pty.
	ptmx, err := pty.Start(command)
	if err != nil {
		return err
	}
	// Make sure to close the pty at the end.
	defer func() {
		_ = ptmx.Close()
	}() // Best effort.

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

	stdin_ctx, stdin_cancel := context.WithCancel(ctx)
	stdin := newContextedStdin(stdin_ctx)

	go func() {
		_, _ = io.Copy(ptmx, &stdin)
		log.Println("Stopped copying to ptmx")
	}()
	_, _ = io.Copy(os.Stdout, ptmx)
	stdin_cancel()
	os.Stdin.Write([]byte{0x04})

	log.Println("Stopped copying to stdout")

	return nil
}

type contextedStdin struct {
	ctx       context.Context
	data_chan chan []byte
}

func newContextedStdin(ctx context.Context) contextedStdin {
	data_chan := make(chan []byte)
	in := contextedStdin{
		ctx:       ctx,
		data_chan: data_chan,
	}
	go func() {
		var buffer [2048]byte
		var is_finished bool
		for {
			n, err := os.Stdin.Read(buffer[:])
			if err != nil {
				return
			}
			for _, c := range buffer {
				if c == 0x04 { // EOF Ctrl+D
					is_finished = true
				}
			}
			select {
			case data_chan <- buffer[:n]:
			case <-ctx.Done():
				return
			}
			if is_finished {
				return
			}
		}
	}()
	return in
}

func (in *contextedStdin) Read(p []byte) (n int, err error) {
	select {
	case new_data := <-in.data_chan:
		copy(p, new_data)
		return len(new_data), nil
	case <-in.ctx.Done():
		return 0, io.EOF
	}
}
