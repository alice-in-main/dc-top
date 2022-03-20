package window

import (
	"fmt"
	"log"

	"github.com/gdamore/tcell/v2"
)

var _screen tcell.Screen = nil

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
	_screen = screen
}

func GetScreen() tcell.Screen {
	return _screen
}

func CloseScreen() {
	if _screen != nil {
		_screen.Clear()
		_screen.Fini()
	} else {
		panic("Tried to close nil screen")
	}
}
