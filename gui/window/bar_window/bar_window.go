package bar_window

import (
	"context"
	"dc-top/gui/window"
	"log"
	"time"

	"github.com/gdamore/tcell/v2"
)

type BarWindow struct {
	context       context.Context
	resize_ch     chan interface{}
	message_chan  chan BarMessage
	enable_toggle chan bool
}

func NewBarWindow(context context.Context) BarWindow {
	return BarWindow{
		context:       context,
		resize_ch:     make(chan interface{}),
		message_chan:  make(chan BarMessage),
		enable_toggle: make(chan bool),
	}
}

func (w *BarWindow) Open() {
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

func (w *BarWindow) Close() {}

func (w *BarWindow) main() {
	x1, y1, x2, y2 := window.ContainersBarWindowSize()
	var state = barState{
		window_state: window.NewWindow(x1, y1, x2, y2),
		message:      _emptyMessage{},
	}
	drawBar(state)

	clear_timer := time.NewTicker(5 * time.Second)
	defer clear_timer.Stop()
	var should_clear int

	for {
		select {
		case <-w.resize_ch:
			x1, y1, x2, y2 := window.ContainersBarWindowSize()
			state.window_state.SetBorders(x1, y1, x2, y2)
			drawBar(state)
		case message_event := <-w.message_chan:
			state.message = message_event
			if should_clear < 5 {
				should_clear++
			}
			drawBar(state)
		case is_enabled := <-w.enable_toggle:
			state.is_enabled = is_enabled
		case <-clear_timer.C:
			if should_clear == 0 {
				state.message = _emptyMessage{}
				drawBar(state)
			} else {
				should_clear--
			}
		case <-w.context.Done():
			log.Printf("Bar window stopped drwaing...\n")
			return
		}
	}
}
