package window

import (
	"context"
	"dc-top/docker/compose"
	"dc-top/gui/elements"
	"log"

	"github.com/gdamore/tcell/v2"
)

type GeneralInfoWindow struct {
	context   context.Context
	resize_ch chan interface{}
}

func NewGeneralInfoWindow(context context.Context) GeneralInfoWindow {
	return GeneralInfoWindow{
		context:   context,
		resize_ch: make(chan interface{}),
	}
}

func (w *GeneralInfoWindow) Open(s tcell.Screen) {
	go w.main(s)
}

func (w *GeneralInfoWindow) Resize() {
	w.resize_ch <- nil
}

func (w *GeneralInfoWindow) KeyPress(_ tcell.EventKey) {}

func (w *GeneralInfoWindow) MousePress(_ tcell.EventMouse) {}

func (w *GeneralInfoWindow) HandleEvent(interface{}, WindowType) (interface{}, error) {
	log.Println("General info window got event")
	panic(1)
}

func (w *GeneralInfoWindow) Close() {}

type generalInfoState struct {
	window_state   WindowState
	version        string
	dc_mode_status string
}

func (w *GeneralInfoWindow) main(s tcell.Screen) {
	x1, y1, x2, y2 := GeneralInfoWindowSize(s)
	state := generalInfoState{
		window_state:   NewWindow(x1, y1, x2, y2),
		version:        "dc-top v0.1",
		dc_mode_status: getDcModeStatus(),
	}
	drawGeneralInfo(s, state)
	for {
		select {
		case <-w.resize_ch:
			x1, y1, x2, y2 := GeneralInfoWindowSize(s)
			state.window_state.SetBorders(x1, y1, x2, y2)
			drawGeneralInfo(s, state)
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

func drawGeneralInfo(screen tcell.Screen, state generalInfoState) {
	DrawContents(screen, &state.window_state, generalInfoDrawerGenerator(&state))
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
