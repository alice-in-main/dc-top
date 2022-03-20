package general_info_window

import (
	"context"
	"dc-top/gui/view/window"
	"log"

	"github.com/gdamore/tcell/v2"
)

type GeneralInfoWindow struct {
	window_ctx    context.Context
	window_cancel context.CancelFunc

	dimensions_generator func() window.Dimensions
	resize_ch            chan interface{}
	enable_toggle        chan bool
}

func NewGeneralInfoWindow() GeneralInfoWindow {
	return GeneralInfoWindow{
		dimensions_generator: func() window.Dimensions {
			x1, y1, x2, y2 := window.GeneralInfoWindowSize()
			return window.NewDimensions(x1, y1, x2, y2, false)
		},
		resize_ch:     make(chan interface{}),
		enable_toggle: make(chan bool),
	}
}

func (w *GeneralInfoWindow) Open(view_ctx context.Context) {
	w.window_ctx, w.window_cancel = context.WithCancel(view_ctx)
	go w.main()
}

func (w *GeneralInfoWindow) Resize() {
	w.resize_ch <- nil
}

func (w *GeneralInfoWindow) KeyPress(_ tcell.EventKey) {}

func (w *GeneralInfoWindow) MousePress(_ tcell.EventMouse) {}

func (w *GeneralInfoWindow) HandleEvent(interface{}, window.WindowType) (interface{}, error) {
	panic(1)
}

func (w *GeneralInfoWindow) Disable() {
	log.Printf("Disable GeneralInfoWindow...")
	w.enable_toggle <- false
}

func (w *GeneralInfoWindow) Enable() {
	log.Printf("Enable GeneralInfoWindow...")
	w.enable_toggle <- true
}

func (w *GeneralInfoWindow) Close() {
	w.window_cancel()
}

func (w *GeneralInfoWindow) main() {
	state := generalInfoState{
		is_enabled:     true,
		version:        "dc-top v0.1",
		dc_mode_status: getDcModeStatus(),
	}
	w.drawGeneralInfo(state)
	for {
		select {
		case <-w.resize_ch:
			w.drawGeneralInfo(state)
		case is_enabled := <-w.enable_toggle:
			state.is_enabled = is_enabled
			if is_enabled {
				w.drawGeneralInfo(state)
			}
		case <-w.window_ctx.Done():
			log.Printf("General info window stopped drwaing...\n")
			return
		}
	}
}

func (w *GeneralInfoWindow) drawGeneralInfo(state generalInfoState) {
	if state.is_enabled {
		dimensions := w.dimensions_generator()
		window.DrawContents(&dimensions, generalInfoDrawerGenerator(&state, window.Width(&dimensions)))
	}
}
