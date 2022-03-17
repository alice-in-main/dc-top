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
	context              context.Context
	resize_ch            chan interface{}
	message_chan         chan BarMessage
	enable_toggle        chan bool
}

func NewBarWindow(context context.Context, dimensions_generator func() window.Dimensions) BarWindow {
	return BarWindow{
		dimensions_generator: dimensions_generator,
		context:              context,
		resize_ch:            make(chan interface{}),
		message_chan:         make(chan BarMessage),
		enable_toggle:        make(chan bool),
	}
}

func (w *BarWindow) Open() {
	go w.main()
}

func (w *BarWindow) Dimensions() window.Dimensions {
	return w.dimensions_generator()
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
			w.drawBar(state)
		case message_event := <-w.message_chan:
			state.message = message_event
			if should_clear < 5 {
				should_clear++
			}
			w.drawBar(state)
		case is_enabled := <-w.enable_toggle:
			state.is_enabled = is_enabled
		case <-clear_timer.C:
			if should_clear == 0 {
				state.message = _emptyMessage{}
				w.drawBar(state)
			} else {
				should_clear--
			}
		case <-w.context.Done():
			log.Printf("Bar window stopped drwaing...\n")
			return
		}
	}
}

func (w *BarWindow) drawBar(state barState) {
	if state.is_enabled {
		tmp_dimensions := w.dimensions_generator()
		window.DrawContents(&tmp_dimensions, generateBarDrawer(state))
	}
}
