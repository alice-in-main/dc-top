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
	id           string
	logs_writer  *logsWriter
	logs_context context.Context
	logs_cancel  context.CancelFunc
}

func NewContainerLogsWindow(id string) ContainerLogsWindow {
	return ContainerLogsWindow{
		id:          id,
		logs_writer: nil,
	}
}

func (w *ContainerLogsWindow) Open(view_ctx context.Context) {
	w.logs_context, w.logs_cancel = context.WithCancel(view_ctx)
	go func() {
		logs_writer := newLogsWriter(w.logs_context)
		w.logs_writer = &logs_writer
		go docker.StreamContainerLogs(w.id, &logs_writer, w.logs_context, w.logs_cancel)
		logs_writer.logPrinter()
		log.Println("Switcing back...")
		logs_writer.drawer_semaphore.Acquire(logs_writer.ctx, 1)
		logs_writer.drawer_semaphore.Release(1)
		window.GetScreen().PostEvent(window.NewReturnUpperViewEvent())
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

func (w *ContainerLogsWindow) Disable() {
	w.logs_writer.enable_toggle <- false
}

func (w *ContainerLogsWindow) Close() { w.logs_cancel() }

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
		w.logs_cancel()
	case tcell.KeyRune:
		switch ev.Rune() {
		case 'h':
			window.GetScreen().PostEvent(window.NewChangeToLogsHelpEvent())
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
			w.logs_cancel()
		case 'l':
			w.logs_cancel()
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
