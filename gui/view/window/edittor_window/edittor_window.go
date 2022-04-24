package edittor_window

import (
	"context"
	"dc-top/gui/view/window"
	"dc-top/gui/view/window/bar_window"
	"fmt"
	"log"
	"os"

	"github.com/gdamore/tcell/v2"
	"golang.org/x/sync/semaphore"
)

type EdittorWindow struct {
	window_context       context.Context
	window_cancel        context.CancelFunc
	dimensions_generator func() window.Dimensions
	drawer_semaphore     *semaphore.Weighted

	file *os.File

	resize_chan   chan interface{}
	enable_toggle chan bool
	keyboard_chan chan *tcell.EventKey
}

func NewEdittorWindow(file *os.File) EdittorWindow {

	return EdittorWindow{
		drawer_semaphore: semaphore.NewWeighted(1),
		dimensions_generator: func() window.Dimensions {
			w, h := window.GetScreen().Size()
			return window.NewDimensions(0, 0, w-1, h-2, false)
		},
		file:          file,
		resize_chan:   make(chan interface{}),
		enable_toggle: make(chan bool),
		keyboard_chan: make(chan *tcell.EventKey),
	}
}

func (w *EdittorWindow) Open(view_context context.Context) {
	w.window_context, w.window_cancel = context.WithCancel(view_context)
	if err := w.main(); err != nil {
		window.GetScreen().PostEvent(window.NewReturnUpperViewEvent())
		bar_window.Err([]rune(fmt.Sprintf("Failed to open file %s for editting", w.file.Name())))
	}
}

func (w *EdittorWindow) Resize() {
	w.resize_chan <- nil
}

func (w *EdittorWindow) KeyPress(ev tcell.EventKey) {
	w.keyboard_chan <- &ev
}

func (w *EdittorWindow) MousePress(_ tcell.EventMouse) {}

func (w *EdittorWindow) HandleEvent(ev interface{}, wt window.WindowType) (interface{}, error) {
	panic("how'd i get here")
}

func (w *EdittorWindow) Disable() {
	w.enable_toggle <- false
	w.drawer_semaphore.Acquire(w.window_context, 1)
	w.drawer_semaphore.Release(1)
}

func (w *EdittorWindow) Enable() {
	w.enable_toggle <- true
}

func (w *EdittorWindow) Close() {
	log.Println("closing")
	w.drawer_semaphore.Acquire(w.window_context, 1)
	defer w.drawer_semaphore.Release(1)
	w.window_cancel()
	w.file.Close()
	log.Println("closed")
}
