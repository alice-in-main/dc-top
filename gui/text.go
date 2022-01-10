package gui

import (
	"github.com/gdamore/tcell/v2"
)

func TextDrawer(str string, style tcell.Style) stringStyler {
	return func(i int) (rune, tcell.Style) {
		if i < len(str) {
			return rune(str[i]), style
		} else {
			return ' ', tcell.StyleDefault
		}
	}
}
