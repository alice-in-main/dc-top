package window

import (
	"context"
	"dc-top/gui/elements"
	"log"
	"time"

	"github.com/gdamore/tcell/v2"
)

type BarWindow struct {
	context      context.Context
	resize_ch    chan interface{}
	message_chan chan BarMessage
}

func NewBarWindow(context context.Context) BarWindow {
	return BarWindow{
		context:      context,
		resize_ch:    make(chan interface{}),
		message_chan: make(chan BarMessage),
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

func (w *BarWindow) HandleEvent(ev interface{}, wt WindowType) (interface{}, error) {
	switch ev := ev.(type) {
	case BarMessage:
		w.message_chan <- ev
	default:
		log.Fatal("Got unknown event in bar", ev)
	}
	return nil, nil
}

func (w *BarWindow) Close() {}

func (w *BarWindow) main(s tcell.Screen) {
	x1, y1, x2, y2 := ContainersBarWindowSize(s)
	var state = barState{
		window_state: NewWindow(x1, y1, x2, y2),
		message:      _emptyMessage{},
	}
	drawBar(s, state)

	clear_timer := time.NewTicker(5 * time.Second)
	defer clear_timer.Stop()
	var should_clear int

	for {
		select {
		case <-w.resize_ch:
			x1, y1, x2, y2 := ContainersBarWindowSize(s)
			state.window_state.SetBorders(x1, y1, x2, y2)
			drawBar(s, state)
		case message_event := <-w.message_chan:
			state.message = message_event
			if should_clear < 5 {
				should_clear++
			}
			drawBar(s, state)
		case <-clear_timer.C:
			if should_clear == 0 {
				state.message = _emptyMessage{}
				drawBar(s, state)
			} else {
				should_clear--
			}
		case <-w.context.Done():
			log.Printf("Bar window stopped drwaing...\n")
			return
		}
	}
}

type barState struct {
	window_state WindowState
	message      BarMessage
}

func drawBar(screen tcell.Screen, state barState) {
	DrawContents(screen, &state.window_state, generateBarDrawer(screen, state))
}

func generateBarDrawer(screen tcell.Screen, state barState) func(x, y int) (rune, tcell.Style) {
	return func(x, y int) (rune, tcell.Style) {
		if y == 0 {
			return state.message.Styler()(x)
		}
		return '\x00', tcell.StyleDefault
	}
}

type BarMessage interface {
	Styler() elements.StringStyler
}

type _emptyMessage struct{}

func (_emptyMessage) Styler() elements.StringStyler {
	return elements.EmptyDrawer()
}

type infoMessage struct {
	msg []rune
}

func (m infoMessage) Styler() elements.StringStyler {
	const info_prefix = "Info: "
	return elements.TextDrawer(info_prefix, tcell.StyleDefault.Foreground(tcell.ColorGreen)).Concat(len(info_prefix), elements.RuneDrawer(m.msg, tcell.StyleDefault))
}

// type warnMessage struct {
// 	msg []rune
// }

// func (m warnMessage) Styler() elements.StringStyler {
// 	var warn_prefix = []rune("\u26a0 Warn: ")
// 	return elements.RuneDrawer(warn_prefix, tcell.StyleDefault.Foreground(tcell.ColorYellow)).Concat(len(warn_prefix), elements.RuneDrawer(m.msg, tcell.StyleDefault))
// }

type errorMessage struct {
	msg []rune
}

func (m errorMessage) Styler() elements.StringStyler {
	var warn_prefix = []rune("\u26a0 Error: ")
	return elements.RuneDrawer(warn_prefix, tcell.StyleDefault.Foreground(tcell.ColorRed)).Concat(len(warn_prefix), elements.RuneDrawer(m.msg, tcell.StyleDefault))
}

// type criticalMessage struct {
// 	msg []rune
// }

// func (m *criticalMessage) Styler() elements.StringStyler {
// 	const crit_prefix = "!!!: "
// 	return elements.TextDrawer(crit_prefix, tcell.StyleDefault.Foreground(tcell.ColorRed)).Concat(len(crit_prefix), elements.RuneDrawer(m.msg, tcell.StyleDefault))
// }
