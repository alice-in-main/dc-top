package error_window

import (
	"context"
	"dc-top/gui/elements"
	"dc-top/gui/view/window"
	"errors"
	"log"

	"github.com/acarl005/stripansi"
	"github.com/gdamore/tcell/v2"
)

type ErrorWindow struct {
	window_ctx    context.Context
	window_cancel context.CancelFunc

	message              []byte
	dimensions_generator func() window.Dimensions
	resize_ch            chan interface{}
	is_enabled           bool

	y_offset int
}

func NewErrorWindow(message []byte) ErrorWindow {
	return ErrorWindow{
		message: message,
		dimensions_generator: func() window.Dimensions {
			x1, y1, x2, y2 := window.ErrorWindowSize()
			return window.NewDimensions(x1, y1, x2, y2, true)
		},
		resize_ch:  make(chan interface{}),
		is_enabled: true,
		y_offset:   0,
	}
}

func (w *ErrorWindow) Open(view_ctx context.Context) {
	log.Println("Opening error")
	w.window_ctx, w.window_cancel = context.WithCancel(view_ctx)
	w.drawError()
}

func (w *ErrorWindow) Resize() {
	w.drawError()
}

func (w *ErrorWindow) KeyPress(ev tcell.EventKey) {
	key := ev.Key()
	switch key {
	case tcell.KeyRune:
		switch ev.Rune() {
		case 'g':
			w.y_offset = 0
			w.drawError()
			return
		case 'q':
			goto to_default
		case ' ':
			goto to_default
		}
		return
	case tcell.KeyCtrlC:
		break
	case tcell.KeyCtrlD:
		break
	case tcell.KeyEscape:
		break
	case tcell.KeyEnter:
		break
	case tcell.KeyUp:
		if w.y_offset > 0 {
			w.y_offset--
			w.drawError()
		}
		return
	case tcell.KeyDown:
		w.y_offset++
		w.drawError()
		return
	default:
		return
	}
to_default:
	window.GetScreen().PostEvent(window.NewReturnUpperViewEvent())
}

func (w *ErrorWindow) MousePress(_ tcell.EventMouse) {}

func (w *ErrorWindow) HandleEvent(interface{}, window.WindowType) (interface{}, error) {
	window.ExitIfErr(errors.New("shouldn't have gotten here"))
	panic(1)
}

func (w *ErrorWindow) Disable() {
	log.Printf("Disable ErrorWindow...")
	w.is_enabled = false
}

func (w *ErrorWindow) Enable() {
	log.Printf("Enable ErrorWindow...")
	w.is_enabled = true
	w.drawError()
}

func (w *ErrorWindow) Close() {
	w.window_cancel()
}

func (w *ErrorWindow) drawError() {
	if !w.is_enabled {
		return
	}
	clean_message := []byte(stripansi.Strip(string(w.message)))
	dimensions := w.dimensions_generator()
	log.Print(string(clean_message))

	header := elements.TextDrawer("Error", tcell.StyleDefault.Bold(true).Underline(true))
	message_lines := make([][]byte, 0)
	for prev_end, i := 0, 0; i < len(clean_message); i++ {
		if clean_message[i] == '\n' || i-prev_end >= window.Width(&dimensions) {
			if clean_message[i] == '\n' {
				i++
			}
			message_lines = append(message_lines, clean_message[prev_end:i])
			prev_end = i
		}
	}

	if window.Height(&dimensions) > len(message_lines)+3 {
		dimensions.ButtomY = dimensions.TopY + len(message_lines) + 3
	}

	var drawer = func(x, y int) (rune, tcell.Style) {
		if y == 0 {
			return header(x)
		}
		if y == 1 {
			return ' ', tcell.StyleDefault
		}
		line_index := y - 2 + w.y_offset
		if line_index < len(message_lines) {
			if x < len(message_lines[line_index]) {
				return rune(message_lines[line_index][x]), tcell.StyleDefault
			}
		}
		return '\x00', tcell.StyleDefault
	}

	window.DrawContents(&dimensions, drawer)
	window.GetScreen().Show()
}
