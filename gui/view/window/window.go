package window

import (
	"context"

	"github.com/gdamore/tcell/v2"
)

type Window interface {
	Open(view_ctx context.Context)
	Resize()
	KeyPress(tcell.EventKey)
	MousePress(tcell.EventMouse)
	HandleEvent(event interface{}, sender WindowType) (interface{}, error)
	Enable()
	Disable()
	Close()
}
