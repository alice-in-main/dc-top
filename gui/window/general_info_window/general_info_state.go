package general_info_window

import (
	"dc-top/docker/compose"
	"dc-top/gui/elements"
	"dc-top/gui/window"
	"log"

	"github.com/gdamore/tcell/v2"
)

type generalInfoState struct {
	window_state   window.WindowState
	is_enabled     bool
	version        string
	dc_mode_status string
}

func (w *GeneralInfoWindow) main() {
	x1, y1, x2, y2 := window.GeneralInfoWindowSize()
	state := generalInfoState{
		window_state:   window.NewWindow(x1, y1, x2, y2),
		is_enabled:     true,
		version:        "dc-top v0.1",
		dc_mode_status: getDcModeStatus(),
	}
	drawGeneralInfo(state)
	for {
		select {
		case <-w.resize_ch:
			x1, y1, x2, y2 := window.GeneralInfoWindowSize()
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

func drawGeneralInfo(state generalInfoState) {
	if state.is_enabled {
		window.DrawContents(&state.window_state, generalInfoDrawerGenerator(&state))
	}
}

func generalInfoDrawerGenerator(state *generalInfoState) func(x, y int) (rune, tcell.Style) {
	header_drawer := elements.TextDrawer(state.dc_mode_status, tcell.StyleDefault).Concat(len(state.dc_mode_status),
		elements.RhsTextDrawer(state.version, tcell.StyleDefault, window.Width(&state.window_state)-len(state.dc_mode_status)))
	return func(x, y int) (rune, tcell.Style) {
		if y == 0 {
			return header_drawer(x)
		}
		return '\x00', tcell.StyleDefault
	}
}

func getDcModeStatus() string {
	if compose.DcModeEnabled() {
		return "Docker Compose mode is enabled."
	} else {
		return "Docker Compose mode is disabled, showing all dockerd containers."
	}
}
