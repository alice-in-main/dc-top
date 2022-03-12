package window

import (
	"context"
	docker "dc-top/docker"
	"dc-top/gui/elements"
	"dc-top/utils"
	"log"

	"github.com/gdamore/tcell/v2"
)

type logsWriter struct {
	ctx          context.Context
	is_following bool
	is_searching bool
	is_enabled   bool
	search_box   elements.TextBox

	logs         [docker.MaxSavedLogs]singleLog
	logs_offset  int
	logs_counter int
	view_offset  int
	lines        []elements.StringStyler

	redraw_request chan interface{}
	write_queue    chan []string
	// pause          chan interface{}
	// resume         chan interface{}
	enable_toggle chan bool
	stop          chan interface{}
}

func newLogsWriter(ctx context.Context) logsWriter {
	_, height := GetScreen().Size()
	new_writer := logsWriter{
		is_following: true,
		is_searching: false,
		is_enabled:   true,
		logs_offset:  0,
		logs_counter: 0,
		view_offset:  0,
		ctx:          ctx,
		search_box: elements.NewTextBox(
			elements.TextDrawer("/ ", tcell.StyleDefault.Foreground(tcell.ColorGreenYellow)),
			2,
			tcell.StyleDefault,
			tcell.StyleDefault.Underline(true)),
		lines:          make([]elements.StringStyler, height),
		redraw_request: make(chan interface{}),
		write_queue:    make(chan []string),
		enable_toggle:  make(chan bool),
		// pause:          make(chan interface{}),
		// resume:         make(chan interface{}),
		stop: make(chan interface{}),
	}
	return new_writer
}

func (writer *logsWriter) Write(logs_batch []byte) (int, error) {
	var nl_index int
	logs := make([]string, 0)
	for offset := 0; nl_index != -1 && offset < len(logs_batch); offset += (nl_index + 1) {
		nl_index = utils.FindByte('\n', []byte(logs_batch[offset:]))
		var log_line string
		if nl_index != -1 {
			log_line = string(logs_batch[offset : offset+nl_index])
		} else {
			log_line = string(logs_batch[offset:])
		}
		logs = append(logs, log_line)
	}
	writer.write_queue <- logs
	return len(logs_batch), nil
}

func (writer *logsWriter) logPrinter() {
	for {
		select {
		case logs := <-writer.write_queue:
			writer.writeLogs(logs)
		case <-writer.redraw_request:
			writer.redraw()
		// case <-writer.pause:
		// 	<-writer.resume
		case is_enabled := <-writer.enable_toggle:
			writer.is_enabled = is_enabled
		case <-writer.ctx.Done():
			return
		}
	}
}

func (writer *logsWriter) writeLogs(logs []string) {
	for _, l := range logs {
		writer.saveLog(l)
	}
	if writer.is_following {
		writer.redraw()
	}
}

func (writer *logsWriter) redraw() {
	writer.updateLines()
	writer.showLines()
}

func (writer *logsWriter) saveLog(_log string) {
	const metadata_len = 8
	if len(_log) > metadata_len {
		metadata := _log[:metadata_len]
		log_line_text := _log[metadata_len:]
		is_stdout := (metadata[0] == 1)
		writer.logs[writer.logs_offset] = newLog(log_line_text, is_stdout)
	} else {
		writer.logs[writer.logs_offset] = newLog("", true)
	}
	writer.logs_offset = (writer.logs_offset + 1) % docker.MaxSavedLogs
	writer.logs_counter++
	if writer.is_following {
		writer.view_offset = writer.logs_counter
	}
}

func (writer *logsWriter) updateLines() {
	width, height := GetScreen().Size()
	writer.lines = make([]elements.StringStyler, height)
	log_i := (writer.view_offset - 1) % docker.MaxSavedLogs
	if log_i < 0 {
		log_i = docker.MaxSavedLogs + log_i
	}
	for line_i := height - 1; line_i >= 0; {
		log_line_text := writer.logs[log_i].content
		log_line_style := tcell.StyleDefault
		if !writer.logs[log_i].is_stdout {
			log_line_style = log_line_style.Foreground(tcell.ColorRed)
		}
		num_partitions := 1 + len(log_line_text)/width
		log_with_highlights := elements.HighlightDrawer(log_line_text, writer.search_box.Value(), log_line_style)
		for j := 0; j < num_partitions; j++ {
			writer.lines[line_i] = elements.Suffix(log_with_highlights, (num_partitions-j-1)*width)
			line_i--
			if line_i < 0 {
				break
			}
		}
		log_i--
		if log_i < 0 {
			log_i = docker.MaxSavedLogs - 1
		}
	}
}

func (writer *logsWriter) showLines() {
	screen := GetScreen()
	screen.Clear()
	width, height := screen.Size()
	for j, line := range writer.lines {
		for i := 0; i < width; i++ {
			if line != nil {
				r, s := line(i)
				screen.SetContent(i, j, r, nil, s)
			}
		}
	}
	if height >= 1 && writer.is_searching {
		search_box := writer.search_box.Style()
		for i := 0; i < width; i++ {
			r, s := search_box(i)
			screen.SetContent(i, 0, r, nil, s)
		}
	} else if height >= 1 && !writer.is_following {
		not_following_message := elements.TextDrawer("Currently not following logs. Press 'f' to start following.", tcell.StyleDefault.Background(tcell.ColorGreen).Bold(true))
		for i := 0; i < width; i++ {
			r, s := not_following_message(i)
			screen.SetContent(i, 0, r, nil, s)
		}
	}
	screen.Sync()
}

func (writer *logsWriter) logStopper(cancel context.CancelFunc) error {
	for {
		select {
		case <-writer.stop:
			log.Println("stopped log stopper from stop chan")
			cancel()
			return nil
		case <-writer.ctx.Done():
			log.Println("stopped log stopper from context")
			cancel()
			return nil
		}
	}
}

type ContainerLogsWindow struct {
	id          string
	logs_writer *logsWriter
	context     context.Context
	cancel      context.CancelFunc
}

func NewContainerLogsWindow(id string) ContainerLogsWindow {
	container_log_window_context, cancel := context.WithCancel(context.TODO())
	return ContainerLogsWindow{
		id:          id,
		logs_writer: nil,
		context:     container_log_window_context,
		cancel:      cancel,
	}
}

func (w *ContainerLogsWindow) Open() {
	go func() {
		container_log_window_context, cancel := context.WithCancel(context.TODO())
		logs_writer := newLogsWriter(container_log_window_context)
		w.logs_writer = &logs_writer
		go logs_writer.logPrinter()
		go func() {
			err := logs_writer.logStopper(cancel)
			exitIfErr(err)
		}()
		go docker.StreamContainerLogs(w.id, &logs_writer, container_log_window_context, cancel)
		<-container_log_window_context.Done()
		log.Println("Switcing back...")
		GetScreen().PostEvent(NewChangeToDefaultViewEvent())
	}()
}

func (w *ContainerLogsWindow) Resize() {
	w.triggerRedraw()
}

func (w *ContainerLogsWindow) KeyPress(ev tcell.EventKey) {
	if w.logs_writer.is_searching {
		w.handleSearchKeyPress(&ev)
	} else {
		w.handleRegularKeyPress(&ev)
	}
}

func (w *ContainerLogsWindow) MousePress(tcell.EventMouse) {
	panic("unimplemented MousePress for logs window")
}

func (w *ContainerLogsWindow) HandleEvent(event interface{}, sender WindowType) (interface{}, error) {
	panic("unimplemented HandleEvent for logs window")
}

func (w *ContainerLogsWindow) Enable() { w.logs_writer.enable_toggle <- true }

func (w *ContainerLogsWindow) Disable() { w.logs_writer.enable_toggle <- false }

func (w *ContainerLogsWindow) Close() { w.cancel() }

func (w *ContainerLogsWindow) handleRegularKeyPress(ev *tcell.EventKey) {
	key := ev.Key()
	switch key {
	case tcell.KeyEnter:
		w.startFollowing()
	case tcell.KeyUp:
		log.Printf("view_offset: %d, logs_counter: %d. %d >= (%d - %d)", w.logs_writer.view_offset, w.logs_writer.logs_counter, w.logs_writer.view_offset, w.logs_writer.logs_counter, docker.MaxSavedLogs)
		if w.logs_writer.view_offset >= (w.logs_writer.logs_counter - docker.MaxSavedLogs) {
			w.logs_writer.view_offset--
			w.logs_writer.is_following = false
			w.triggerRedraw()
		}
	case tcell.KeyDown:
		if w.logs_writer.view_offset < w.logs_writer.logs_counter {
			w.logs_writer.view_offset++
			w.logs_writer.is_following = false
			w.triggerRedraw()
		}
	case tcell.KeyCtrlD:
		w.logs_writer.stop <- nil
	case tcell.KeyRune:
		switch ev.Rune() {
		case 't':
			for i, _log := range w.logs_writer.logs {
				log.Print(i, _log)
			}
		case 'f':
			w.startFollowing()
		case '/':
			w.logs_writer.is_searching = true
			w.triggerRedraw()
		case 'c':
			w.logs_writer.search_box.Reset()
			w.triggerRedraw()
		case 'q':
			w.logs_writer.stop <- nil
		case 'l':
			w.logs_writer.stop <- nil
		}
	}
}

func (w *ContainerLogsWindow) handleSearchKeyPress(ev *tcell.EventKey) {
	key := ev.Key()
	switch key {
	case tcell.KeyCtrlD:
		w.logs_writer.search_box.Reset()
		w.logs_writer.is_searching = false
	case tcell.KeyEscape:
		w.logs_writer.search_box.Reset()
		w.logs_writer.is_searching = false
	case tcell.KeyEnter:
		w.logs_writer.is_searching = false
	default:
		w.logs_writer.search_box.HandleKey(ev)
	}
	w.triggerRedraw()
}

func (w *ContainerLogsWindow) startFollowing() {
	w.logs_writer.is_following = true
	w.logs_writer.view_offset = w.logs_writer.logs_counter
	w.triggerRedraw()
}

func (w *ContainerLogsWindow) triggerRedraw() {
	w.logs_writer.redraw_request <- nil
}

type singleLog struct {
	content   string
	is_stdout bool
}

func newLog(content string, is_stdout bool) singleLog {
	return singleLog{
		content:   content,
		is_stdout: is_stdout,
	}
}
