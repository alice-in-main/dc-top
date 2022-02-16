package window

import (
	"context"
	"dc-top/docker"
	"dc-top/gui/elements"
	"dc-top/utils"
	"log"

	"github.com/gdamore/tcell/v2"
)

type ContainerLogWindow struct {
	id          string
	resize_chan chan interface{}
	stop_chan   chan interface{}
}

type logWindowState struct {
	window_state WindowState
}

type logsWriter struct {
	logs              [docker.NumSavedLogs][]byte
	screen            tcell.Screen
	cached_state      logWindowState
	state_updates     chan logWindowState
	inner_write_index int
}

func (writer *logsWriter) Write(logs_batch []byte) (int, error) {
	var nl_index int
	for offset := 0; nl_index != -1 && offset < len(logs_batch); offset += (nl_index + 1) {
		nl_index = utils.FindByte('\n', []byte(logs_batch[offset:]))
		if nl_index != -1 {
			writer.logs[writer.inner_write_index] = logs_batch[offset : offset+nl_index]
		} else {
			writer.logs[writer.inner_write_index] = logs_batch[offset:]
		}
		log_line := writer.logs[writer.inner_write_index]
		writer.writeSingleLogLine(log_line)
		writer.inner_write_index = (writer.inner_write_index + 1) % docker.NumSavedLogs
	}
	return len(logs_batch), nil
}

func (writer *logsWriter) writeSingleLogLine(log_line []byte) {
	const metadata_len = 8
	log_line_metadata := log_line[:metadata_len]
	var log_line_text string
	if len(log_line) > metadata_len {
		log_line_text = string(log_line[metadata_len:])
	} else {
		log_line_text = "<empty>"
	}
	is_stdout := log_line_metadata[0] == 1
	line_style := tcell.StyleDefault
	if !is_stdout {
		line_style = line_style.Background(tcell.ColorDarkOrchid)
	}
	line := elements.TextDrawer(log_line_text, line_style)
	for i := 0; i < 100; i++ {
		r, s := line(i)
		log.Print(string(r))
		writer.screen.SetContent(i, writer.inner_write_index, r, nil, s)
	}
	writer.screen.Show()
}

func newLogsWriter(screen tcell.Screen) logsWriter {
	return logsWriter{
		inner_write_index: 0,
		screen:            screen,
	}
}

func NewContainerLogWindow(id string) ContainerLogWindow {
	return ContainerLogWindow{
		id:          id,
		resize_chan: make(chan interface{}),
		stop_chan:   make(chan interface{}),
	}
}

func (w *ContainerLogWindow) Open(s tcell.Screen) {
	s.DisableMouse()
	go w.main(s)
}

func (w *ContainerLogWindow) Resize() {
	w.resize_chan <- nil
}

func (w *ContainerLogWindow) KeyPress(key tcell.EventKey) {}

func (w *ContainerLogWindow) MousePress(ev tcell.EventMouse) {}

func (w *ContainerLogWindow) Close() {
	w.stop_chan <- nil
}

func (w *ContainerLogWindow) main(screen tcell.Screen) {
	width, height := screen.Size()
	state := logWindowState{
		window_state: NewWindow(screen, 0, 0, width, height),
	}
	logs_writer := newLogsWriter(screen)
	container_log_window_context, cancel := context.WithCancel(context.Background())
	go docker.StreamContainerLogs(w.id, &logs_writer, container_log_window_context)
	for {
		select {
		case <-w.resize_chan:
			width, height = screen.Size()
			state.window_state.SetBorders(0, 0, width, height)
		case <-w.stop_chan:
			cancel()
			return
		}
	}
}
