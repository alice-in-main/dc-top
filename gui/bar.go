package gui

import (
	"github.com/gdamore/tcell/v2"
)

func BarDrawer(min_val float64, max_val float64, curr_val float64, bar_len int) stringStyler {
	normalized_max := max_val - min_val
	normalized_curr := curr_val - min_val
	notch := (max_val - min_val) / float64(bar_len)
	loading_bar_rune := '\u2584'
	return func(i int) (rune, tcell.Style) {
		switch {
		case i > bar_len-1:
			return ' ', tcell.StyleDefault
		case float64(i)*notch >= normalized_curr:
			return loading_bar_rune, tcell.StyleDefault.Foreground(tcell.ColorDarkGray)
		case float64(i)*notch >= 0.80*(normalized_max):
			return loading_bar_rune, tcell.StyleDefault.Foreground(tcell.ColorRed)
		case float64(i)*notch >= 0.50*(normalized_max):
			return loading_bar_rune, tcell.StyleDefault.Foreground(tcell.ColorYellow)
		case float64(i)*notch >= 0:
			return loading_bar_rune, tcell.StyleDefault.Foreground(tcell.ColorGreen)
		}
		panic("Illegal bar state")
	}
}
