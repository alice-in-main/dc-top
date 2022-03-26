package container_logs_window

import (
	"context"
	docker "dc-top/docker"
	"dc-top/gui/elements"
	"dc-top/gui/view/window"
	"dc-top/gui/view/window/bar_window"
	"dc-top/utils"
	"fmt"
	"log"

	"github.com/gdamore/tcell/v2"
	"golang.org/x/sync/semaphore"
)

type logsWriter struct {
	ctx              context.Context
	drawer_semaphore *semaphore.Weighted

	is_following         bool
	is_typing            bool
	is_enabled           bool
	is_looking           bool
	dimensions_generator func() window.Dimensions

	logs_container LogContainer
	logs_offset    int
	logs_counter   int
	view_offset    int
	lines          []elements.StringStyler

	search_box     elements.TextBox
	lookup_request chan interface{}
	next_search    chan interface{}
	prev_search    chan interface{}

	redraw_request chan interface{}
	write_queue    chan []string
	enable_toggle  chan bool
}

func newLogsWriter(ctx context.Context) logsWriter {
	_, y1, _, y2 := window.LogsWindowSize()
	height := y2 - y1
	logs_container := NewArrStringSearcher(docker.MaxSavedLogs)
	new_writer := logsWriter{
		ctx:              ctx,
		drawer_semaphore: semaphore.NewWeighted(1),

		is_following: true,
		is_typing:    false,
		is_enabled:   true,
		is_looking:   false,
		dimensions_generator: func() window.Dimensions {
			x1, y1, x2, y2 := window.LogsWindowSize()
			return window.NewDimensions(x1, y1, x2, y2, false)
		},

		logs_container: &logs_container,
		logs_offset:    0,
		logs_counter:   0,
		view_offset:    0,
		lines:          make([]elements.StringStyler, height),

		search_box: elements.NewTextBox(
			elements.TextDrawer("/ ", tcell.StyleDefault.Foreground(tcell.ColorGreenYellow)),
			2,
			tcell.StyleDefault,
			tcell.StyleDefault.Underline(true),
			true),
		lookup_request: make(chan interface{}),
		next_search:    make(chan interface{}),
		prev_search:    make(chan interface{}),

		redraw_request: make(chan interface{}),
		write_queue:    make(chan []string),
		enable_toggle:  make(chan bool),
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
	writer.drawer_semaphore.Acquire(writer.ctx, 1)
	defer writer.drawer_semaphore.Release(1)
	for {
		select {
		case <-writer.lookup_request:
			if writer.search_box.Value() == "" {
				break
			}
			writer.is_following = false
			indices := writer.logs_container.Search(writer.search_box.Value())
			bar_window.Info([]rune(fmt.Sprintf("Found %d results for %s", len(indices), writer.search_box.Value())))
			if len(indices) > 0 {
				writer.handleLookup(indices)
			}
			writer.redraw()
		case logs := <-writer.write_queue:
			writer.writeLogs(logs)
		case <-writer.redraw_request:
			if writer.is_enabled {
				writer.redraw()
			}
		case is_enabled := <-writer.enable_toggle:
			writer.is_enabled = is_enabled
			if writer.is_enabled {
				writer.redraw()
			}
		case <-writer.ctx.Done():
			return
		}
	}
}

func (writer *logsWriter) handleLookup(result_indices []int) {
	writer.is_looking = true
	defer func() { writer.is_looking = false }()
	i := 0
	writer.view_offset = result_indices[i]
	writer.redraw()
	for {
		select {
		case <-writer.next_search:
			if i == len(result_indices)-1 {
				i = 0
			} else {
				i++
			}
		case <-writer.prev_search:
			if i == 0 {
				i = len(result_indices) - 1
			} else {
				i--
			}
		case is_enabled := <-writer.enable_toggle:
			writer.is_enabled = is_enabled
		case <-writer.redraw_request:
			log.Printf("Exitting lookup from redraw")
			return
		case <-writer.ctx.Done():
			log.Printf("Exitting lookup from context")
			return
		}
		if writer.is_enabled {
			bar_window.Info([]rune(fmt.Sprintf("Showing result %d/%d", i+1, len(result_indices))))
			writer.view_offset = result_indices[i]
			writer.redraw()
		}
	}
}

func (writer *logsWriter) writeLogs(logs []string) {
	for _, l := range logs {
		writer.saveLog(l)
	}
	if writer.is_following && writer.is_enabled {
		writer.redraw()
	}
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
		writer.view_offset = writer.logs_counter - 1
	}
}

func (writer *logsWriter) redraw() {
	writer.updateLines()
	writer.showLines()
}

func (writer *logsWriter) updateLines() {
	dimensions := writer.dimensions_generator()
	width := window.Width(&dimensions)
	height := window.Height(&dimensions)
	writer.lines = make([]elements.StringStyler, height)
	log_i := writer.view_offset % docker.MaxSavedLogs
	if log_i < 0 {
		log_i = docker.MaxSavedLogs + log_i
	}
	for line_i := height - 1; line_i >= 0; {
		log_line_text := writer.logs_container.Get(log_i).content
		log_line_style := tcell.StyleDefault
		if !writer.logs_container.Get(log_i).is_stdout {
			log_line_style = log_line_style.Foreground(tcell.ColorRed)
		}
		num_partitions := 1 + (len(log_line_text)-1)/width
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
	dimensions := writer.dimensions_generator()
	height := window.Height(&dimensions)
	search_box := writer.search_box.Style()
	not_following_message := elements.TextDrawer("Currently not following logs. Press 'f' to start following.", tcell.StyleDefault.Background(tcell.ColorGreen).Bold(true))
	log_drawer := func(i int, j int) (rune, tcell.Style) {
		if j == 0 {
			if height >= 1 && writer.is_typing {
				return search_box(i)
			} else if height >= 1 && !writer.is_following {
				return not_following_message(i)
			}
		}
		if writer.lines[j] != nil {
			return writer.lines[j](i)
		}

		return '\x00', tcell.StyleDefault
	}
	window.DrawContents(&dimensions, log_drawer)
	window.GetScreen().Sync()
}
