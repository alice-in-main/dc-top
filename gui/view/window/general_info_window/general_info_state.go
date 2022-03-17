package general_info_window

import (
	"dc-top/docker/compose"
	"dc-top/gui/elements"

	"github.com/gdamore/tcell/v2"
)

type generalInfoState struct {
	is_enabled     bool
	version        string
	dc_mode_status string
}

func generalInfoDrawerGenerator(state *generalInfoState, width int) func(x, y int) (rune, tcell.Style) {
	header_drawer := elements.TextDrawer(state.dc_mode_status, tcell.StyleDefault).Concat(len(state.dc_mode_status),
		elements.RhsTextDrawer(state.version, tcell.StyleDefault, width-len(state.dc_mode_status)))
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
