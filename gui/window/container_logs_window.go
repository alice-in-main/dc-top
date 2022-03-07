package window

import (
	"context"
	docker "dc-top/docker"
	"dc-top/gui/elements"
	"dc-top/utils"
	"fmt"
	"log"
	"regexp"

	"github.com/gdamore/tcell/v2"
)

type searchMode uint8

const (
	str searchMode = iota
	regex
)

type logsWriter struct {
	screen       tcell.Screen
	search_mode  searchMode
	curr_re      *regexp.Regexp
	search_query string
	ctx          context.Context

	logs        [docker.NumSavedLogs]string
	logs_offset int
	lines       []elements.StringStyler

	regex_search_chan  chan string
	string_search_chan chan string
	write_queue        chan []string
	pause              chan interface{}
	resume             chan interface{}
	stop               chan interface{}
}

func newLogsWriter(screen tcell.Screen, ctx context.Context) logsWriter {
	_, height := screen.Size()
	new_writer := logsWriter{
		screen:             screen,
		search_query:       "",
		logs_offset:        0,
		ctx:                ctx,
		lines:              make([]elements.StringStyler, height),
		regex_search_chan:  make(chan string),
		string_search_chan: make(chan string),
		write_queue:        make(chan []string),
		pause:              make(chan interface{}),
		resume:             make(chan interface{}),
		stop:               make(chan interface{}),
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
		case search := <-writer.regex_search_chan:
			writer.search_mode = regex
			writer.search_query = search
			re, err := regexp.Compile(writer.search_query)
			if err != nil {
				fmt.Println("Failed to compile regex ", writer.search_query)
				break
			}
			writer.curr_re = re
		case search := <-writer.string_search_chan:
			writer.search_mode = str
			writer.search_query = search
		case logs := <-writer.write_queue:
			writer.writeLogs(logs)
		case <-writer.pause:
			<-writer.resume
		case <-writer.ctx.Done():
			return
		}
	}
}

func (writer *logsWriter) writeLogs(logs []string) {
	// const metadata_len = 8
	// log_line_metadata := single_log[:metadata_len]
	// var log_line_text
	// if len(single_log) > metadata_len {
	// 	log_line_text = string(single_log[metadata_len:])
	// }
	// is_stdout := log_line_metadata[0] == 1
	// if !is_stdout {
	// 	log_line_text = elements.Foreground(log_line_text, elements.Purple)
	// }
	// if writer.search_query != "" {
	// 	if writer.search_mode == regex {
	// 		log_line_text = writer.curr_re.ReplaceAllStringFunc(log_line_text,
	// 			func(s string) string { return elements.Background(s, elements.B_DarkGray) })
	// 	} else if writer.search_mode == str {
	// 		log_line_text = strings.ReplaceAll(log_line_text, writer.search_query, elements.Background(writer.search_query, elements.B_DarkGray))
	// 	}

	// }
	for _, l := range logs {
		writer.saveLog(l)
	}
	writer.updateLines()
	writer.showLines()
}

func (writer *logsWriter) saveLog(_log string) {
	writer.logs[writer.logs_offset] = _log
	writer.logs_offset = (writer.logs_offset + 1) % docker.NumSavedLogs
}

func (writer *logsWriter) updateLines() {
	const metadata_len = 8
	width, height := writer.screen.Size()
	writer.lines = make([]elements.StringStyler, height)
	log_i := writer.logs_offset - 1
	for line_i := 0; line_i < height; {
		curr_log := writer.logs[log_i]
		log_i--
		if log_i < 0 {
			log_i = docker.NumSavedLogs - 1
		}
		if len(curr_log) > metadata_len {
			log_line_text := curr_log[metadata_len:]
			cut_log := utils.CutString([]byte(log_line_text), width)
			for j, line := range cut_log {
				reverse_index := height - line_i + j - len(cut_log)
				if reverse_index < 0 {
					break
				}
				writer.lines[reverse_index] = elements.TextDrawer(string(line), tcell.StyleDefault)
			}
			line_i += len(cut_log)
		} else {
			line_i++
		}
	}
}

func (writer *logsWriter) showLines() {
	width, _ := writer.screen.Size()
	for j, line := range writer.lines {
		for i := 0; i < width; i++ {
			if line != nil {
				r, s := line(i)
				writer.screen.SetContent(i, j, r, nil, s)
			}
		}
	}
	writer.screen.Sync()
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

	// err := keyboard.Open()
	// if err != nil {
	// 	return err
	// }
	// defer keyboard.Close()
	// for {
	// 	char, key, err := keyboard.GetSingleKey()
	// 	if err != nil && !strings.HasPrefix(err.Error(), "Unrecognized escape sequence") {
	// 		log.Printf("Got error while waiting for key '%s'", err.Error())
	// 		cancel()
	// 	}
	// 	if key == keyboard.KeyEsc || key == keyboard.KeyCtrlC || char == 'q' || char == 'l' {
	// 		cancel()
	// 		// TODO: graceful exit
	// 		if key == keyboard.KeyCtrlC {
	// 			os.Exit(0)
	// 		}
	// 		return nil
	// 	} else if char == '/' || char == '?' {
	// 		writer.pause <- nil
	// 		search_reader := bufio.NewReader(os.Stdin)
	// 		fmt.Print(string(char), " Enter search: ")
	// 		text, err := search_reader.ReadString('\n')
	// 		writer.resume <- nil
	// 		if err != nil {
	// 			log.Println("Failed to read user search", err)
	// 			cancel()
	// 		}
	// 		text = text[:len(text)-1] // remove newline char
	// 		if text != "" {
	// 			if char == '?' {
	// 				writer.regex_search_chan <- text
	// 			} else {
	// 				writer.string_search_chan <- text
	// 			}
	// 		}
	// 		log.Println("ready for new search ")
	// 	} else {
	// 		continue
	// 	}
	// }
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

func (w *ContainerLogsWindow) Open(screen tcell.Screen) {
	go func() {
		_, height := screen.Size()
		for i := 0; i < height; i++ {
			fmt.Println()
		}
		container_log_window_context, cancel := context.WithCancel(context.TODO())
		logs_writer := newLogsWriter(screen, container_log_window_context)
		w.logs_writer = &logs_writer
		go logs_writer.logPrinter()
		go func() {
			err := logs_writer.logStopper(cancel)
			exitIfErr(screen, err)
		}()
		go docker.StreamContainerLogs(w.id, &logs_writer, container_log_window_context, cancel)
		<-container_log_window_context.Done()
		log.Println("Switcing back...")
		screen.PostEvent(NewChangeToDefaultViewEvent())
	}()
}

func (w *ContainerLogsWindow) Resize() {
	w.logs_writer.write_queue <- []string{}
}

func (w *ContainerLogsWindow) KeyPress(ev tcell.EventKey) {
	key := ev.Key()
	switch key {
	case tcell.KeyCtrlD:
		w.logs_writer.stop <- nil
	case tcell.KeyRune:
		switch ev.Rune() {
		case 'q':
			w.logs_writer.stop <- nil
		case 'l':
			w.logs_writer.stop <- nil
		}
	}
}

func (w *ContainerLogsWindow) MousePress(tcell.EventMouse) {
	panic("unimplemented MousePress for logs window")
}

func (w *ContainerLogsWindow) HandleEvent(event interface{}, sender WindowType) (interface{}, error) {
	panic("unimplemented HandleEvent for logs window")
}

func (w *ContainerLogsWindow) Enable() { w.logs_writer.resume <- nil }

func (w *ContainerLogsWindow) Disable() { w.logs_writer.pause <- nil }

func (w *ContainerLogsWindow) Close() { w.cancel() }
