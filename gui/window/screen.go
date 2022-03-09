package window

import "github.com/gdamore/tcell/v2"

var _screen tcell.Screen = nil

func InitScreen(screen tcell.Screen) {
	if _screen == nil {
		_screen = screen
	} else {
		panic("Tried to reinitiate screen")
	}
}

func GetScreen() tcell.Screen {
	return _screen
}
