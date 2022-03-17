package bar_window

import (
	"dc-top/gui/elements"

	"github.com/gdamore/tcell/v2"
)

type barState struct {
	message    BarMessage
	is_enabled bool
}

var bar_prefix = elements.TextDrawer("> ", tcell.StyleDefault.Foreground(tcell.ColorYellow))

func generateBarDrawer(state barState) func(x, y int) (rune, tcell.Style) {
	return func(x, y int) (rune, tcell.Style) {
		if y == 0 {
			return bar_prefix.Concat(2, state.message.Styler())(x)
		}
		return '\x00', tcell.StyleDefault
	}
}
