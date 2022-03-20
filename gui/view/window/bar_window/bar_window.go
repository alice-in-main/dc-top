package bar_window

import (
	"context"
	"dc-top/gui/view/window"
	"log"
	"time"

	"github.com/gdamore/tcell/v2"
)

type BarWindow struct {
	dimensions_generator func() window.Dimensions
	window_context       context.Context
	window_cancel        context.CancelFunc

	resize_ch     chan interface{}
	message_chan  chan BarMessage
	enable_toggle chan bool
}

func NewBarWindow(dimensions_generator func() window.Dimensions) BarWindow {
	return BarWindow{
		dimensions_generator: dimensions_generator,
		resize_ch:            make(chan interface{}),
		message_chan:         make(chan BarMessage),
		enable_toggle:        make(chan bool),
	}
}

func (w *BarWindow) Open(view_context context.Context) {
	w.window_context, w.window_cancel = context.WithCancel(view_context)
	go w.main()
}

func (w *BarWindow) Resize() {
	w.resize_ch <- nil
}

func (w *BarWindow) KeyPress(ev tcell.EventKey) {}

func (w *BarWindow) MousePress(_ tcell.EventMouse) {}

func (w *BarWindow) HandleEvent(ev interface{}, wt window.WindowType) (interface{}, error) {
	switch ev := ev.(type) {
	case BarMessage:
		w.message_chan <- ev
	default:
		log.Fatal("Got unknown event in bar", ev)
	}
	return nil, nil
}

func (w *BarWindow) Disable() {
	log.Printf("Disable bar...")
	w.enable_toggle <- false
}

func (w *BarWindow) Enable() {
	log.Printf("Enable bar...")
	w.enable_toggle <- true
}

func (w *BarWindow) Close() {
	w.window_cancel()
}

func (w *BarWindow) main() {
	var state = barState{
		message:    _emptyMessage{},
		is_enabled: true,
	}
	w.drawBar(state)

	clear_timer := time.NewTicker(5 * time.Second)
	defer clear_timer.Stop()
	var should_clear int

	for {
		select {
		case <-w.resize_ch:
		case message_event := <-w.message_chan:
			state.message = message_event
			if should_clear < 5 {
				should_clear++
			}
		case is_enabled := <-w.enable_toggle:
			state.is_enabled = is_enabled
		case <-clear_timer.C:
			if should_clear == 0 {
				state.message = _emptyMessage{}
			} else {
				should_clear--
			}
		case <-w.window_context.Done():
			log.Printf("Bar window stopped drwaing...\n")
			return
		}
		w.drawBar(state)
		window.GetScreen().Show()
	}
}

func (w *BarWindow) drawBar(state barState) {
	if state.is_enabled {
		tmp_dimensions := w.dimensions_generator()
		window.DrawContents(&tmp_dimensions, generateBarDrawer(state))
	}
}
