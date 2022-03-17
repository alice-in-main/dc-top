package general_info_window

import (
	"context"
	"dc-top/gui/view/window"
	"log"

	"github.com/gdamore/tcell/v2"
)

type GeneralInfoWindow struct {
	dimensions    window.Dimensions
	context       context.Context
	resize_ch     chan interface{}
	enable_toggle chan bool
}

func NewGeneralInfoWindow(context context.Context) GeneralInfoWindow {
	x1, y1, x2, y2 := window.GeneralInfoWindowSize()
	return GeneralInfoWindow{
		dimensions:    window.NewDimensions(x1, y1, x2, y2, false),
		context:       context,
		resize_ch:     make(chan interface{}),
		enable_toggle: make(chan bool),
	}
}

func (w *GeneralInfoWindow) Open() {
	go w.main()
}

func (w *GeneralInfoWindow) Dimensions() window.Dimensions {
	return w.dimensions
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

func (w *GeneralInfoWindow) Close() {}

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
			x1, y1, x2, y2 := window.GeneralInfoWindowSize()
			w.dimensions.SetBorders(x1, y1, x2, y2)
			w.drawGeneralInfo(state)
		case is_enabled := <-w.enable_toggle:
			state.is_enabled = is_enabled
			if is_enabled {
				w.drawGeneralInfo(state)
			}
		case <-w.context.Done():
			log.Printf("General info window stopped drwaing...\n")
			return
		}
	}
}

func (w *GeneralInfoWindow) drawGeneralInfo(state generalInfoState) {
	if state.is_enabled {
		window.DrawContents(&w.dimensions, generalInfoDrawerGenerator(&state, window.Width(&w.dimensions)))
	}
}
