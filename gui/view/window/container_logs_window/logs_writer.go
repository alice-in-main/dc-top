package container_logs_window

import (
	"context"
	docker "dc-top/docker"
	"dc-top/gui/elements"
	"dc-top/gui/view/window"
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
	dimensions   window.Dimensions

	logs_container LogContainer
	logs_offset    int
	logs_counter   int
	view_offset    int
	lines          []elements.StringStyler

	redraw_request chan interface{}
	write_queue    chan []string
	enable_toggle  chan bool
	stop           chan interface{}
}

func newLogsWriter(ctx context.Context) logsWriter {
	x1, y1, x2, y2 := window.LogsWindowSize()
	height := y2 - y1
	logs_container := NewArrStringSearcher(docker.MaxSavedLogs)
	new_writer := logsWriter{
		is_following:   true,
		is_searching:   false,
		is_enabled:     true,
		dimensions:     window.NewDimensions(x1, y1, x2, y2, false),
		logs_container: &logs_container,
		logs_offset:    0,
		logs_counter:   0,
		view_offset:    0,
		ctx:            ctx,
		search_box: elements.NewTextBox(
			elements.TextDrawer("/ ", tcell.StyleDefault.Foreground(tcell.ColorGreenYellow)),
			2,
			tcell.StyleDefault,
			tcell.StyleDefault.Underline(true)),
		lines:          make([]elements.StringStyler, height),
		redraw_request: make(chan interface{}),
		write_queue:    make(chan []string),
		enable_toggle:  make(chan bool),
		stop:           make(chan interface{}),
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
	var new_log containerLog
	if len(_log) > metadata_len {
		metadata := _log[:metadata_len]
		log_line_text := _log[metadata_len:]
		is_stdout := (metadata[0] == 1)
		new_log = newLog(log_line_text, is_stdout)
	} else {
		new_log = newLog("", true)
	}
	writer.logs_container.Put(&new_log, writer.logs_offset)
	writer.logs_offset = (writer.logs_offset + 1) % docker.MaxSavedLogs
	writer.logs_counter++
	if writer.is_following {
		writer.view_offset = writer.logs_counter
	}
}

func (writer *logsWriter) updateLines() {
	width := window.Width(&writer.dimensions)
	height := window.Height(&writer.dimensions)
	writer.lines = make([]elements.StringStyler, height)
	log_i := (writer.view_offset - 1) % docker.MaxSavedLogs
	if log_i < 0 {
		log_i = docker.MaxSavedLogs + log_i
	}
	for line_i := height - 1; line_i >= 0; {
		log_line_text := writer.logs_container.Get(log_i).content
		log_line_style := tcell.StyleDefault
		if !writer.logs_container.Get(log_i).is_stdout {
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
	height := window.Height(&writer.dimensions)
	log_drawer := func(i int, j int) (rune, tcell.Style) {
		if j == 0 {
			if height >= 1 && writer.is_searching {
				search_box := writer.search_box.Style()
				return search_box(i)
			} else if height >= 1 && !writer.is_following {
				not_following_message := elements.TextDrawer("Currently not following logs. Press 'f' to start following.", tcell.StyleDefault.Background(tcell.ColorGreen).Bold(true))
				return not_following_message(i)
			}
		}
		if writer.lines[j] != nil {
			return writer.lines[j](i)
		}

		return '\x00', tcell.StyleDefault
	}
	window.DrawContents(&writer.dimensions, log_drawer)
	window.GetScreen().Sync()
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
