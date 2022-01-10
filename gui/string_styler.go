package gui

import "github.com/gdamore/tcell/v2"

type stringStyler func(x int) (rune, tcell.Style)
