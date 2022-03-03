package window

import (
	"bufio"
	"context"
	docker "dc-top/docker"
	"dc-top/gui/elements"
	"dc-top/utils"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/eiannone/keyboard"
	"github.com/gdamore/tcell/v2"
)

type searchMode uint8

const (
	str searchMode = iota
	regex
)

type logsWriter struct {
	screen             tcell.Screen
	search_mode        searchMode
	curr_re            *regexp.Regexp
	regex_search_chan  chan string
	string_search_chan chan string
	search_query       string
	write_queue        chan []byte
	pause              chan interface{}
	resume             chan interface{}
}

func newLogsWriter(screen tcell.Screen) logsWriter {
	new_writer := logsWriter{
		screen:             screen,
		regex_search_chan:  make(chan string),
		string_search_chan: make(chan string),
		search_query:       "",
		write_queue:        make(chan []byte),
		pause:              make(chan interface{}),
		resume:             make(chan interface{}),
	}
	return new_writer
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
		writer.write_queue <- log_line
	}
	return len(logs_batch), nil
}

func (writer *logsWriter) logPrinter(context context.Context) {
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
		case log := <-writer.write_queue:
			writer.writeSingleLog(log)
		case <-writer.pause:
			<-writer.resume
		case <-context.Done():
			return
		}
	}
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
		log_line_text = elements.Foreground(log_line_text, elements.Purple)
	}
	if writer.search_query != "" {
		if writer.search_mode == regex {
			log_line_text = writer.curr_re.ReplaceAllStringFunc(log_line_text,
				func(s string) string { return elements.Background(s, elements.B_DarkGray) })
		} else if writer.search_mode == str {
			log_line_text = strings.ReplaceAll(log_line_text, writer.search_query, elements.Background(writer.search_query, elements.B_DarkGray))
		}

	}
	fmt.Println(string(log_line_text))
}

func (writer *logsWriter) logStopper(cancel context.CancelFunc) {
	for {
		char, key, err := keyboard.GetSingleKey()
		if err != nil && !strings.HasPrefix(err.Error(), "Unrecognized escape sequence") {
			log.Printf("Got error while waiting for key '%s'", err.Error())
			cancel()
		}
		if key == keyboard.KeyEsc || key == keyboard.KeyCtrlC || char == 'q' || char == 'l' {
			cancel()
			if key == keyboard.KeyCtrlC {
				os.Exit(0)
			}
			return
		} else if char == '/' || char == '?' {
			writer.pause <- nil
			search_reader := bufio.NewReader(os.Stdin)
			fmt.Print(string(char), " Enter search: ")
			text, err := search_reader.ReadString('\n')
			writer.resume <- nil
			if err != nil {
				log.Println("Failed to read user search", err)
				cancel()
			}
			text = text[:len(text)-1] // remove newline
			log.Println("searcing for ", text)
			if text != "" {
				if char == '?' {
					writer.regex_search_chan <- text
				} else {
					writer.string_search_chan <- text
				}
			}
			log.Println("ready for new search ")
		} else {
			continue
		}
	}
}

type ContainerLogWindow struct {
	id      string
	context context.Context
	cancel  context.CancelFunc
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

func (w *ContainerLogWindow) HandleEvent(interface{}, WindowType) (interface{}, error) {
	log.Println("Log window got event")
	panic(1)
}

func (w *ContainerLogWindow) Close() {
	w.cancel()
}

func (w *ContainerLogWindow) main(screen tcell.Screen) {
	screen.Suspend()
	_, height := screen.Size()
	for i := 0; i < height; i++ {
		fmt.Println()
	}
	err := keyboard.Open()
	exitIfErr(screen, err)
	defer keyboard.Close()
	container_log_window_context, cancel := context.WithCancel(context.TODO())
	logs_writer := newLogsWriter(screen)
	go logs_writer.logPrinter(container_log_window_context)
	go logs_writer.logStopper(cancel)
	go docker.StreamContainerLogs(w.id, &logs_writer, container_log_window_context, cancel)
	<-container_log_window_context.Done()
	screen.Resume()
	log.Println("Switcing back...")
	screen.PostEvent(NewChangeToDefaultViewEvent())
}
