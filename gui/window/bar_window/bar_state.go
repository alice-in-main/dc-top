package bar_window

import (
	"dc-top/gui/window"

	"github.com/gdamore/tcell/v2"
)

type barState struct {
	window_state window.WindowState
	message      BarMessage
	is_enabled   bool
}

func drawBar(state barState) {
	if state.is_enabled {
		window.DrawContents(&state.window_state, generateBarDrawer(state))
	}
}

func generateBarDrawer(state barState) func(x, y int) (rune, tcell.Style) {
	return func(x, y int) (rune, tcell.Style) {
		if y == 0 {
			return state.message.Styler()(x)
		}
		return '\x00', tcell.StyleDefault
	}
}
