package subshell_window

import (
	"bufio"
	"context"
	"dc-top/gui/elements"
	"dc-top/gui/view/window"

	"github.com/gdamore/tcell/v2"
)

type SubshellWindow struct {
	dimensions_generator func() window.Dimensions
	window_context       context.Context
	window_cancel        context.CancelFunc
	subshell             bufio.Reader
}

func NewSubshellWindow(input_from_subshell bufio.Reader) SubshellWindow {
	return SubshellWindow{
		dimensions_generator: func() window.Dimensions {
			w, h := window.GetScreen().Size()
			return window.NewDimensions(0, w-1, 0, h-1, false)
		},
		subshell: input_from_subshell,
	}
}

func (w *SubshellWindow) Open(view_context context.Context) {
	w.window_context, w.window_cancel = context.WithCancel(view_context)
	go func() {
		for {
			select {
			case <-w.window_context.Done():
				return
			default:
			}
			dimensions := w.dimensions_generator()
			rows := make([]elements.StringStyler, window.Height(&dimensions))
			for y := 0; y < window.Height(&dimensions); y++ {
				var buff [1024]byte
				n, _ := w.subshell.Read(buff[:])
				rows[y] = elements.TextDrawer(string(buff[:n]), tcell.StyleDefault)
			}
			subshell_drawer := func(x, y int) (rune, tcell.Style) {
				return rows[y](x)
			}
			window.DrawContents(&dimensions, subshell_drawer)
		}
	}()
}

func (w *SubshellWindow) Resize() {
}

func (w *SubshellWindow) KeyPress(ev tcell.EventKey) {}

func (w *SubshellWindow) MousePress(_ tcell.EventMouse) {}

func (w *SubshellWindow) HandleEvent(ev interface{}, wt window.WindowType) (interface{}, error) {
	panic("how'd i get here")
}

func (w *SubshellWindow) Disable() {
}

func (w *SubshellWindow) Enable() {
}

func (w *SubshellWindow) Close() {
	w.window_cancel()
}
