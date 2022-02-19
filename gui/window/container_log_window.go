package window

import (
	"context"
	"dc-top/docker"
	"dc-top/gui/elements"
	"dc-top/gui/gui_events"
	"dc-top/utils"
	"fmt"
	"log"

	"github.com/eiannone/keyboard"
	"github.com/gdamore/tcell/v2"
)

type ContainerLogWindow struct {
	id      string
	context context.Context
	cancel  context.CancelFunc
}
type logsWriter struct {
	draw_queue chan []byte
	screen     tcell.Screen
}

func (writer *logsWriter) Write(logs_batch []byte) (int, error) {
	var nl_index int
	for offset := 0; nl_index != -1 && offset < len(logs_batch); offset += (nl_index + 1) {
		nl_index = utils.FindByte('\n', []byte(logs_batch[offset:]))
		var log_line []byte
		if nl_index != -1 {
			log_line = logs_batch[offset : offset+nl_index]
		} else {
			log_line = logs_batch[offset:]
		}
		writer.writeSingleLog(log_line)
	}
	return len(logs_batch), nil
}

func (writer *logsWriter) writeSingleLog(single_log []byte) {
	const metadata_len = 8
	log_line_metadata := single_log[:metadata_len]
	var log_line_text string
	if len(single_log) > metadata_len {
		log_line_text = string(single_log[metadata_len:])
	}
	is_stdout := log_line_metadata[0] == 1
	if !is_stdout {
		log_line_text = elements.Colorize(log_line_text, elements.Red)
	}
	fmt.Println(string(log_line_text))
}

func newLogsWriter(screen tcell.Screen) logsWriter {
	new_writer := logsWriter{
		draw_queue: make(chan []byte),
		screen:     screen,
	}
	return new_writer
}

func NewContainerLogWindow(id string) ContainerLogWindow {
	container_log_window_context, cancel := context.WithCancel(context.TODO())
	return ContainerLogWindow{
		id:      id,
		context: container_log_window_context,
		cancel:  cancel,
	}
}

func (w *ContainerLogWindow) Open(s tcell.Screen) {
	go w.main(s)
}

func (w *ContainerLogWindow) Resize() {}

func (w *ContainerLogWindow) KeyPress(_ tcell.EventKey) {}

func (w *ContainerLogWindow) MousePress(_ tcell.EventMouse) {}

func (w *ContainerLogWindow) Close() {
	w.cancel()
}

func logStopper(cancel context.CancelFunc) {

	if err := keyboard.Open(); err != nil {
		log.Fatal("Failed to start keyboard in container log window")
	}
	defer keyboard.Close()

	for {
		char, key, err := keyboard.GetSingleKey()
		if err != nil {
			log.Fatalf("Got error while waiting for key '%s'", err.Error())
		}
		if key == keyboard.KeyEsc || key == keyboard.KeyCtrlC || char == 'q' || char == 'l' {
			cancel()
			return
		}
	}
}

func (w *ContainerLogWindow) main(screen tcell.Screen) {
	screen.Suspend()
	_, height := screen.Size()
	for i := 0; i < height; i++ {
		fmt.Println()
	}
	container_log_window_context, cancel := context.WithCancel(context.TODO())
	logs_writer := newLogsWriter(screen)
	go logStopper(cancel)
	go docker.StreamContainerLogs(w.id, &logs_writer, container_log_window_context)
	<-container_log_window_context.Done()
	cancel()
	screen.Resume()
	log.Println("Switcing back...")
	screen.PostEvent(gui_events.NewChangeToDefaultViewEvent())
}
