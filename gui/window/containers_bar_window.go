package window

import (
	"context"
	"dc-top/gui/elements"
	"log"

	"github.com/gdamore/tcell/v2"
)

type BarWindow struct {
	context         context.Context
	resize_ch       chan interface{}
	message_updates chan []rune
}

func NewBarWindow(context context.Context) BarWindow {
	return BarWindow{
		context:         context,
		resize_ch:       make(chan interface{}),
		message_updates: make(chan []rune),
	}
}

func (w *BarWindow) Open(s tcell.Screen) {
	go w.main(s)
}

func (w *BarWindow) Resize() {
	w.resize_ch <- nil
}

func (w *BarWindow) KeyPress(ev tcell.EventKey) {}

func (w *BarWindow) MousePress(_ tcell.EventMouse) {}

func (w *BarWindow) HandleEvent(ev interface{}) (interface{}, error) {
	switch ev := ev.(type) {
	case []rune:
		w.message_updates <- ev
	default:
		log.Fatal("Got unknown event")
	}
	return nil, nil
}

func (w *BarWindow) Close() {}

func (w *BarWindow) main(s tcell.Screen) {
	x1, y1, x2, y2 := ContainersBarWindowSize(s)
	var state = barState{
		window_state: NewWindow(s, x1, y1, x2, y2),
		message:      []rune{'\uf600'},
		search:       []rune{},
		index:        0,
	}
	drawBar(state)
	for {
		select {
		case <-w.resize_ch:
			x1, y1, x2, y2 := ContainersBarWindowSize(s)
			state.window_state.SetBorders(x1, y1, x2, y2)
			drawBar(state)
		case message_event := <-w.message_updates:
			state.message = message_event
			drawBar(state)
		case <-w.context.Done():
			log.Printf("Bar window stopped drwaing...\n")
			return
		}
	}
}

type barState struct {
	window_state WindowState
	message      []rune
	search       []rune
	index        int
}

func drawBar(state barState) {
	DrawBorders(&state.window_state, tcell.StyleDefault)
	DrawContents(&state.window_state, generateBarDrawer(state))
}

func generateBarDrawer(state barState) func(x, y int) (rune, tcell.Style) {
	var bar_elements map[int]elements.StringStyler = make(map[int]elements.StringStyler)
	bar_elements[0] = elements.RuneDrawer(state.message, tcell.StyleDefault)
	return func(x, y int) (rune, tcell.Style) {
		if y == 0 {
			return bar_elements[y](x)
		}
		return ' ', tcell.StyleDefault
	}
}
