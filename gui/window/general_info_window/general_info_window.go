package general_info_window

import (
	"context"
	"dc-top/gui/window"
	"log"

	"github.com/gdamore/tcell/v2"
)

type GeneralInfoWindow struct {
	context       context.Context
	resize_ch     chan interface{}
	enable_toggle chan bool
}

func NewGeneralInfoWindow(context context.Context) GeneralInfoWindow {
	return GeneralInfoWindow{
		context:       context,
		resize_ch:     make(chan interface{}),
		enable_toggle: make(chan bool),
	}
}

func (w *GeneralInfoWindow) Open() {
	go w.main()
}

func (w *GeneralInfoWindow) Resize() {
	w.resize_ch <- nil
}

func (w *GeneralInfoWindow) KeyPress(_ tcell.EventKey) {}

func (w *GeneralInfoWindow) MousePress(_ tcell.EventMouse) {}

func (w *GeneralInfoWindow) HandleEvent(interface{}, window.WindowType) (interface{}, error) {
	panic(1)
}

func (w *GeneralInfoWindow) Disable() {
	log.Printf("Disable GeneralInfoWindow...")
	w.enable_toggle <- false
}

func (w *GeneralInfoWindow) Enable() {
	log.Printf("Enable GeneralInfoWindow...")
	w.enable_toggle <- true
}

func (w *GeneralInfoWindow) Close() {}
