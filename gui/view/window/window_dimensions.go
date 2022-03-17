package window

import (
	"log"

	"github.com/gdamore/tcell/v2"
)

type Dimensions struct {
	LeftX    int
	RightX   int
	TopY     int
	ButtomY  int
	Bordered bool
}

func NewDimensions(x1, y1, x2, y2 int, bordered bool) Dimensions {
	if x1 > x2 {
		log.Printf("Bad window coordinates:\ntop left:(%d,%d), buttom right:(%d,%d)\n", x1, y1, x2, y2)
		x1 = x2
	}
	if y1 > y2 {
		log.Printf("Bad window coordinates:\ntop left:(%d,%d), buttom right:(%d,%d)\n", x1, y1, x2, y2)
		y1 = y2
	}
	return Dimensions{
		LeftX:    x1,
		RightX:   x2,
		TopY:     y1,
		ButtomY:  y2,
		Bordered: bordered,
	}
}

func (window_state *Dimensions) RelativeMousePosition(ev *tcell.EventMouse) (int, int) {
	abs_x, abs_y := ev.Position()
	return abs_x - window_state.LeftX, abs_y - window_state.TopY
}

func (window_state *Dimensions) IsOutbounds(ev *tcell.EventMouse) bool {
	abs_x, abs_y := ev.Position()
	return abs_x < window_state.LeftX || abs_x > window_state.RightX || abs_y < window_state.TopY || abs_y > window_state.ButtomY
}

func (window_state *Dimensions) SetBorders(x1, y1, x2, y2 int) {
	window_state.LeftX = x1
	window_state.TopY = y1
	window_state.RightX = x2
	window_state.ButtomY = y2
}

func DrawBorders(window_state *Dimensions) {
	screen := GetScreen()
	style := tcell.StyleDefault.Foreground(tcell.ColorOrangeRed)
	for col := window_state.LeftX; col <= window_state.RightX; col++ {
		screen.SetContent(col, window_state.TopY, tcell.RuneHLine, nil, style)
		screen.SetContent(col, window_state.ButtomY, tcell.RuneHLine, nil, style)
	}
	for row := window_state.TopY; row < window_state.ButtomY; row++ {
		screen.SetContent(window_state.LeftX, row, tcell.RuneVLine, nil, style)
		screen.SetContent(window_state.RightX, row, tcell.RuneVLine, nil, style)
	}

	var ul_corner rune = tcell.RuneULCorner
	var ur_corner rune = tcell.RuneURCorner
	var ll_corner rune = tcell.RuneLLCorner
	var lr_corner rune = tcell.RuneLRCorner

	screen.SetContent(window_state.LeftX, window_state.TopY, ul_corner, nil, style)
	screen.SetContent(window_state.RightX, window_state.TopY, ur_corner, nil, style)
	screen.SetContent(window_state.LeftX, window_state.ButtomY, ll_corner, nil, style)
	screen.SetContent(window_state.RightX, window_state.ButtomY, lr_corner, nil, style)
}

func DrawContents(window_state *Dimensions, contents_generator func(int, int) (rune, tcell.Style)) {
	offset_x := window_state.LeftX
	offset_y := window_state.TopY
	if window_state.Bordered {
		offset_x++
		offset_y++
	}
	screen := GetScreen()
	width := Width(window_state)
	height := Height(window_state)
	for i := 0; i < width; i++ {
		for j := 0; j < height; j++ {
			r, s := contents_generator(i, j)
			screen.SetContent(offset_x+i, offset_y+j, r, nil, s)
		}
	}
}

func Width(window_state *Dimensions) int {
	if window_state.Bordered {
		return window_state.RightX - window_state.LeftX - 1
	}
	return window_state.RightX - window_state.LeftX + 1
}

func Height(window_state *Dimensions) int {
	if window_state.Bordered {
		return window_state.ButtomY - window_state.TopY - 1
	}
	return window_state.ButtomY - window_state.TopY + 1
}
