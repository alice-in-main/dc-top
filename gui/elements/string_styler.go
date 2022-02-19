package elements

import (
	"log"
	"math"

	"github.com/gdamore/tcell/v2"
)

type StringStyler func(x int) (rune, tcell.Style)

func StrikeThrough(orig_styler StringStyler) StringStyler {
	return func(i int) (rune, tcell.Style) {
		r, s := orig_styler(i)
		return r, s.Background(tcell.ColorDarkRed)
	}
}

func TextDrawer(str string, style tcell.Style) StringStyler {
	return func(i int) (rune, tcell.Style) {
		if i < len(str) {
			return rune(str[i]), style
		} else {
			return ' ', tcell.StyleDefault
		}
	}
}

func EmptyDrawer() StringStyler {
	return func(_ int) (rune, tcell.Style) { return ' ', tcell.StyleDefault }
}

func RuneRepeater(r rune, s tcell.Style) StringStyler {
	return func(_ int) (rune, tcell.Style) { return r, s }
}

func PercentageBarDrawer(description string, percentage float64, bar_len int, extra_info []rune) StringStyler {
	var high_percentage float64 = 80.0
	var mid_percentage float64 = 50.0
	var low_percentage float64 = 2.0
	loading_bar_rune := '\u2584'
	desc_len := len(description)
	extra_info_len := len(extra_info)
	return func(i int) (rune, tcell.Style) {
		if i < desc_len {
			return rune(description[i]), tcell.StyleDefault
		}
		bar_percentage := 100.0 * float64(i-desc_len) / float64(bar_len)
		switch {
		case i > bar_len+len(description)-1:
			if i < extra_info_len+bar_len+desc_len {
				return extra_info[i-desc_len-bar_len], tcell.StyleDefault
			}
			return '\x00', tcell.StyleDefault
		case bar_percentage > percentage || percentage < low_percentage:
			return loading_bar_rune, tcell.StyleDefault.Foreground(tcell.ColorDarkGray)
		case percentage >= high_percentage && bar_percentage >= high_percentage:
			return loading_bar_rune, tcell.StyleDefault.Foreground(tcell.ColorRed)
		case percentage >= mid_percentage && bar_percentage >= mid_percentage:
			return loading_bar_rune, tcell.StyleDefault.Foreground(tcell.ColorYellow)
		case percentage >= low_percentage:
			return loading_bar_rune, tcell.StyleDefault.Foreground(tcell.ColorGreen)
		case math.IsNaN(percentage):
			return '\x00', tcell.StyleDefault
		}
		log.Printf("Illegal bar state: got %f percentage and %d bar length\n", percentage, bar_len)
		panic(1)
	}
}

func ValuesBarDrawer(description string, min_val float64, max_val float64, curr_val float64, bar_len int, extra_info []rune) StringStyler {
	normalized_max := max_val - min_val
	normalized_curr := curr_val - min_val
	return PercentageBarDrawer(description, 100.0*normalized_curr/normalized_max, bar_len, extra_info)
}
