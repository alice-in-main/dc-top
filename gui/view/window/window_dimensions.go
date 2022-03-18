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

func (dimensions *Dimensions) RelativeMousePosition(ev *tcell.EventMouse) (int, int) {
	abs_x, abs_y := ev.Position()
	return abs_x - dimensions.LeftX, abs_y - dimensions.TopY
}

func (dimensions *Dimensions) IsOutbounds(ev *tcell.EventMouse) bool {
	abs_x, abs_y := ev.Position()
	return abs_x < dimensions.LeftX || abs_x > dimensions.RightX || abs_y < dimensions.TopY || abs_y > dimensions.ButtomY
}

func (dimensions *Dimensions) SetBorders(x1, y1, x2, y2 int) {
	dimensions.LeftX = x1
	dimensions.TopY = y1
	dimensions.RightX = x2
	dimensions.ButtomY = y2
}

func drawBorders(dimensions *Dimensions) {
	screen := GetScreen()
	style := tcell.StyleDefault.Foreground(tcell.ColorOrangeRed)
	for col := dimensions.LeftX; col <= dimensions.RightX; col++ {
		screen.SetContent(col, dimensions.TopY, tcell.RuneHLine, nil, style)
		screen.SetContent(col, dimensions.ButtomY, tcell.RuneHLine, nil, style)
	}
	for row := dimensions.TopY; row < dimensions.ButtomY; row++ {
		screen.SetContent(dimensions.LeftX, row, tcell.RuneVLine, nil, style)
		screen.SetContent(dimensions.RightX, row, tcell.RuneVLine, nil, style)
	}

	var ul_corner rune = tcell.RuneULCorner
	var ur_corner rune = tcell.RuneURCorner
	var ll_corner rune = tcell.RuneLLCorner
	var lr_corner rune = tcell.RuneLRCorner

	screen.SetContent(dimensions.LeftX, dimensions.TopY, ul_corner, nil, style)
	screen.SetContent(dimensions.RightX, dimensions.TopY, ur_corner, nil, style)
	screen.SetContent(dimensions.LeftX, dimensions.ButtomY, ll_corner, nil, style)
	screen.SetContent(dimensions.RightX, dimensions.ButtomY, lr_corner, nil, style)
}

func DrawContents(dimensions *Dimensions, contents_generator func(int, int) (rune, tcell.Style)) {
	offset_x := dimensions.LeftX
	offset_y := dimensions.TopY
	if dimensions.Bordered {
		offset_x++
		offset_y++
	}
	screen := GetScreen()
	width := Width(dimensions)
	height := Height(dimensions)
	for i := 0; i < width; i++ {
		for j := 0; j < height; j++ {
			r, s := contents_generator(i, j)
			screen.SetContent(offset_x+i, offset_y+j, r, nil, s)
		}
	}

	if dimensions.Bordered {
		drawBorders(dimensions)
	}
}

func Width(dimensions *Dimensions) int {
	if dimensions.Bordered {
		return dimensions.RightX - dimensions.LeftX - 1
	}
	return dimensions.RightX - dimensions.LeftX + 1
}

func Height(dimensions *Dimensions) int {
	if dimensions.Bordered {
		return dimensions.ButtomY - dimensions.TopY - 1
	}
	return dimensions.ButtomY - dimensions.TopY + 1
}
