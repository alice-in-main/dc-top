package window

import (
	"context"
	"dc-top/docker/compose"
	"dc-top/gui/elements"
	"log"

	"github.com/gdamore/tcell/v2"
)

type GeneralInfoWindow struct {
	context       context.Context
	resize_ch     chan interface{}
	enable_toggle chan bool
}

func NewGeneralInfoWindow(context context.Context) GeneralInfoWindow {
	return GeneralInfoWindow{
		context:       context,
		resize_ch:     make(chan interface{}),
		enable_toggle: make(chan bool),
	}
}

func (w *GeneralInfoWindow) Open() {
	go w.main()
}

func (w *GeneralInfoWindow) Resize() {
	w.resize_ch <- nil
}

func (w *GeneralInfoWindow) KeyPress(_ tcell.EventKey) {}

func (w *GeneralInfoWindow) MousePress(_ tcell.EventMouse) {}

func (w *GeneralInfoWindow) HandleEvent(interface{}, WindowType) (interface{}, error) {
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

type generalInfoState struct {
	window_state   WindowState
	is_enabled     bool
	version        string
	dc_mode_status string
}

func (w *GeneralInfoWindow) main() {
	x1, y1, x2, y2 := GeneralInfoWindowSize()
	state := generalInfoState{
		window_state:   NewWindow(x1, y1, x2, y2),
		is_enabled:     true,
		version:        "dc-top v0.1",
		dc_mode_status: getDcModeStatus(),
	}
	drawGeneralInfo(state)
	for {
		select {
		case <-w.resize_ch:
			x1, y1, x2, y2 := GeneralInfoWindowSize()
			state.window_state.SetBorders(x1, y1, x2, y2)
			drawGeneralInfo(state)
		case is_enabled := <-w.enable_toggle:
			state.is_enabled = is_enabled
			if is_enabled {
				drawGeneralInfo(state)
			}
		case <-w.context.Done():
			log.Printf("General info window stopped drwaing...\n")
			return
		}
	}
}

func getDcModeStatus() string {
	if compose.DcModeEnabled() {
		return "Docker Compose mode is enabled."
	} else {
		return "Docker Compose mode is disabled, showing all dockerd containers."
	}
}

func drawGeneralInfo(state generalInfoState) {
	if state.is_enabled {
		DrawContents(&state.window_state, generalInfoDrawerGenerator(&state))
	}
}

func generalInfoDrawerGenerator(state *generalInfoState) func(x, y int) (rune, tcell.Style) {
	header_drawer := elements.TextDrawer(state.dc_mode_status, tcell.StyleDefault).Concat(len(state.dc_mode_status),
		elements.RhsTextDrawer(state.version, tcell.StyleDefault, Width(&state.window_state)-len(state.dc_mode_status)))
	return func(x, y int) (rune, tcell.Style) {
		if y == 0 {
			return header_drawer(x)
		}
		return '\x00', tcell.StyleDefault
	}
}
