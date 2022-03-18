package container_logs_window

import (
	"context"
	docker "dc-top/docker"
	"dc-top/gui/view/window"
	"dc-top/gui/view/window/bar_window"
	"fmt"
	"log"

	"github.com/gdamore/tcell/v2"
)

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
			window.ExitIfErr(err)
		}()
		go docker.StreamContainerLogs(w.id, &logs_writer, container_log_window_context, cancel)
		<-container_log_window_context.Done()
		log.Println("Switcing back...")
		window.GetScreen().PostEvent(window.NewChangeToDefaultViewEvent())
	}()
}

func (w *ContainerLogsWindow) Resize() {
	w.triggerRedraw()
}

func (w *ContainerLogsWindow) KeyPress(ev tcell.EventKey) {
	if w.logs_writer.is_typing {
		w.handleSearchKeyPress(&ev)
	} else {
		w.handleRegularKeyPress(&ev)
	}
}

func (w *ContainerLogsWindow) MousePress(tcell.EventMouse) {
	panic("unimplemented MousePress for logs window")
}

func (w *ContainerLogsWindow) HandleEvent(event interface{}, sender window.WindowType) (interface{}, error) {
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
			w.stopFollowing()
		}
	case tcell.KeyDown:
		if w.logs_writer.view_offset < w.logs_writer.logs_counter {
			w.logs_writer.view_offset++
			w.stopFollowing()
		}
	case tcell.KeyCtrlD:
		w.logs_writer.stop <- nil
	case tcell.KeyRune:
		switch ev.Rune() {
		case 'f':
			w.startFollowing()
		case '/':
			w.logs_writer.is_typing = true
			w.triggerRedraw()
		case 'c':
			w.logs_writer.search_box.Reset()
			bar_window.Info([]rune("Cleared search"))
			w.triggerRedraw()
		case 'n':
			if w.logs_writer.is_looking {
				w.logs_writer.next_search <- nil
			} else {
				w.logs_writer.lookup_request <- nil
			}
		case 'N':
			if w.logs_writer.is_looking {
				w.logs_writer.prev_search <- nil
			} else {
				w.logs_writer.lookup_request <- nil
			}
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
		w.logs_writer.is_typing = false
	case tcell.KeyEscape:
		w.logs_writer.search_box.Reset()
		w.logs_writer.is_typing = false
	case tcell.KeyEnter:
		bar_window.Info([]rune(fmt.Sprintf("Searching for '%s'", w.logs_writer.search_box.Value())))
		w.logs_writer.is_typing = false
	default:
		w.logs_writer.search_box.HandleKey(ev)
	}
	w.triggerRedraw()
}

func (w *ContainerLogsWindow) startFollowing() {
	w.logs_writer.is_following = true
	w.logs_writer.view_offset = w.logs_writer.logs_counter - 1
	bar_window.Info([]rune("Following..."))
	w.triggerRedraw()
}

func (w *ContainerLogsWindow) stopFollowing() {
	w.logs_writer.is_following = false
	bar_window.Info([]rune("Stopped following logs"))
	w.triggerRedraw()
}

func (w *ContainerLogsWindow) triggerRedraw() {
	w.logs_writer.redraw_request <- nil
}
