package elements

import (
	"github.com/gdamore/tcell/v2"
)

func (s1 StringStyler) Concat(stich_index int, s2 StringStyler) StringStyler {
	return func(x int) (r rune, s tcell.Style) {
		if x < stich_index {
			return s1(x)
		} else {
			return s2(x - stich_index)
		}
	}
}
