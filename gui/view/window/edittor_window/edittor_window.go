package edittor_window

import (
	"context"
	"dc-top/gui/view/window"
	"os"

	"github.com/gdamore/tcell/v2"
)

type EdittorWindow struct {
	window_context       context.Context
	window_cancel        context.CancelFunc
	dimensions_generator func() window.Dimensions

	file *os.File

	resize_chan   chan interface{}
	enable_toggle chan bool
	keyboard_chan chan *tcell.EventKey
}

func NewEdittorWindow(file *os.File) EdittorWindow {

	return EdittorWindow{
		dimensions_generator: func() window.Dimensions {
			w, h := window.GetScreen().Size()
			return window.NewDimensions(0, 0, w-1, h-1, false)
		},
		file:          file,
		resize_chan:   make(chan interface{}),
		enable_toggle: make(chan bool),
		keyboard_chan: make(chan *tcell.EventKey),
	}
}

func (w *EdittorWindow) Open(view_context context.Context) {
	w.window_context, w.window_cancel = context.WithCancel(view_context)
	w.main() // TODO: handle error
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
}

func (w *EdittorWindow) Enable() {
	w.enable_toggle <- true
}

func (w *EdittorWindow) Close() {
	w.window_cancel()
}
