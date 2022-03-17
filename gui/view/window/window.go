package window

import (
	"github.com/gdamore/tcell/v2"
)

type Window interface {
	Open()
	Dimensions() Dimensions
	Resize()
	KeyPress(tcell.EventKey)
	MousePress(tcell.EventMouse)
	HandleEvent(event interface{}, sender WindowType) (interface{}, error)
	Enable()
	Disable()
	Close()
}
