package elements

import (
	"fmt"
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
			return '\x00', tcell.StyleDefault
		}
	}
}

func RhsTextDrawer(str string, style tcell.Style, window_width int) StringStyler {
	start_index := window_width - len(str)
	return func(i int) (rune, tcell.Style) {
		if i >= start_index {
			return rune(str[i-start_index]), style
		}
		return '\x00', tcell.StyleDefault
	}
}

func RuneDrawer(str []rune, style tcell.Style) StringStyler {
	return func(i int) (rune, tcell.Style) {
		if i < len(str) {
			return str[i], style
		} else {
			return '\x00', tcell.StyleDefault
		}
	}
}

func IntegerDrawer(n int, style tcell.Style) StringStyler {
	return func(i int) (rune, tcell.Style) {
		str := fmt.Sprintf("%d", n)
		if i < len(str) {
			return rune(str[i]), style
		} else {
			return '\x00', tcell.StyleDefault
		}
	}
}

func EmptyDrawer() StringStyler {
	return func(_ int) (rune, tcell.Style) { return '\x00', tcell.StyleDefault }
}

func RuneRepeater(r rune, s tcell.Style) StringStyler {
	return func(_ int) (rune, tcell.Style) { return r, s }
}

func RuneNRepeater(r rune, n int, s tcell.Style) StringStyler {
	return func(i int) (rune, tcell.Style) {
		if i < n {
			return r, s
		} else {
			return '\x00', tcell.StyleDefault
		}
	}
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

func TextBoxDrawer(text string, cursor_index int, default_style tcell.Style, cursor_style tcell.Style) StringStyler {
	return func(i int) (rune, tcell.Style) {
		var r rune = ' '
		var s tcell.Style = default_style

		if i < len(text) {
			r = rune(text[i])
		}

		if i == cursor_index {
			s = cursor_style
		}

		return r, s
	}
}

func (s1 StringStyler) Concat(stich_index int, s2 StringStyler) StringStyler {
	return func(x int) (r rune, s tcell.Style) {
		if x < stich_index {
			return s1(x)
		} else {
			return s2(x - stich_index)
		}
	}
}
