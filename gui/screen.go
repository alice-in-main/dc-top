package gui

import (
	"dc-top/gui/view/window"
	"fmt"
	"log"

	"github.com/gdamore/tcell/v2"
)

func InitScreen() {
	screen, err := tcell.NewScreen()
	if err != nil {
		log.Printf("%+v", err)
		panic(fmt.Sprintf("%+v", err))
	}
	if err = screen.Init(); err != nil {
		log.Printf("%+v", err)
		panic(fmt.Sprintf("%+v", err))
	}
	window.InitScreen(screen)
}

func CloseScreen() {
	window.CloseScreen()
}
