package help_window

import (
	"context"
	"dc-top/gui/elements"
	"dc-top/gui/view/window"
	"errors"
	"log"

	"github.com/gdamore/tcell/v2"
)

type HelpWindow struct {
	controls []Control

	dimensions_generator func() window.Dimensions
	context              context.Context
	resize_ch            chan interface{}
	is_enabled           bool
}

func NewHelpWindow(context context.Context, controls []Control, dimensions_generator func() window.Dimensions) HelpWindow {
	return HelpWindow{
		controls:             controls,
		dimensions_generator: dimensions_generator,
		context:              context,
		resize_ch:            make(chan interface{}),
		is_enabled:           true,
	}
}

func (w *HelpWindow) Open() {
	w.drawHelp()
}

func (w *HelpWindow) Resize() {
	w.drawHelp()
}

func (w *HelpWindow) KeyPress(ev tcell.EventKey) {
	key := ev.Key()
	switch key {
	case tcell.KeyRune:
		switch ev.Rune() {
		case 'q':
			goto to_default
		case 'h':
			goto to_default
		}
		return
	case tcell.KeyCtrlD:
		break
	case tcell.KeyEscape:
		break
	case tcell.KeyEnter:
		break
	default:
		return
	}
to_default:
	window.GetScreen().PostEvent(window.NewReturnUpperViewEvent())
}

func (w *HelpWindow) MousePress(_ tcell.EventMouse) {}

func (w *HelpWindow) HandleEvent(interface{}, window.WindowType) (interface{}, error) {
	window.ExitIfErr(errors.New("shouldn't have gotten here"))
	panic(1)
}

func (w *HelpWindow) Disable() {
	log.Printf("Disable HelpWindow...")
	w.is_enabled = false
}

func (w *HelpWindow) Enable() {
	log.Printf("Enable HelpWindow...")
	w.is_enabled = true
	w.drawHelp()
}

func (w *HelpWindow) Close() {}

func (w *HelpWindow) drawHelp() {
	if !w.is_enabled {
		return
	}
	var cells = make([][]elements.StringStyler, len(w.controls))
	i := 0
	for _, control := range w.controls {
		cells[i] = []elements.StringStyler{
			elements.TextDrawer(control.key, tcell.StyleDefault),
			elements.TextDrawer(control.meaning, tcell.StyleDefault),
		}
		i++
	}
	var relative_widths = []float64{0.2, 0.8}
	dimensions := w.dimensions_generator()
	var table = []elements.StringStyler{elements.TextDrawer("Controls:", tcell.StyleDefault.Bold(true).Underline(true)),
		elements.EmptyDrawer()}
	table = append(table, elements.TableWithoutSeperator(window.Width(&dimensions), relative_widths, cells)...)
	var drawer = func(x, y int) (rune, tcell.Style) {
		if y >= len(table) {
			return ' ', tcell.StyleDefault
		}
		return table[y](x)
	}
	window.DrawContents(&dimensions, drawer)
	window.GetScreen().Show()
}
