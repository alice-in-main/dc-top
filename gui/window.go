package gui

import (
	"log"

	"github.com/gdamore/tcell/v2"
)

type Window struct {
	screen   tcell.Screen
	left_x   int
	right_x  int
	top_y    int
	buttom_y int
}

func NewWindow(screen tcell.Screen, x1, y1, x2, y2 int) Window {
	if x1 >= x2 || y1 >= y2 {
		log.Printf("Bad window coordinates:\ntop left:(%d,%d), buttom right:(%d,%d)\n", x1, y1, x2, y2)
		panic(1)
	}
	return Window{
		screen:   screen,
		left_x:   x1,
		right_x:  x2,
		top_y:    y1,
		buttom_y: y2,
	}
}

func (window *Window) DrawBorders(style tcell.Style) {
	for col := window.left_x; col <= window.right_x; col++ {
		window.screen.SetContent(col, window.top_y, tcell.RuneHLine, nil, style)
		window.screen.SetContent(col, window.buttom_y, tcell.RuneHLine, nil, style)
	}
	for row := window.top_y; row < window.buttom_y; row++ {
		window.screen.SetContent(window.left_x, row, tcell.RuneVLine, nil, style)
		window.screen.SetContent(window.right_x, row, tcell.RuneVLine, nil, style)
	}

	window.screen.SetContent(window.left_x, window.top_y, tcell.RuneULCorner, nil, style)
	window.screen.SetContent(window.right_x, window.top_y, tcell.RuneURCorner, nil, style)
	window.screen.SetContent(window.left_x, window.buttom_y, tcell.RuneLLCorner, nil, style)
	window.screen.SetContent(window.right_x, window.buttom_y, tcell.RuneLRCorner, nil, style)
}

func (window *Window) DrawContents(contents_generator func(int, int) (rune, tcell.Style)) {
	width := window.right_x - window.left_x - 1
	height := window.buttom_y - window.top_y - 1
	offset_x := window.left_x + 1
	offset_y := window.top_y + 1
	for i := 0; i < width; i++ {
		for j := 0; j < height; j++ {
			r, s := contents_generator(i, j)
			window.screen.SetContent(offset_x+i, offset_y+j, r, nil, s)
		}
	}
}
