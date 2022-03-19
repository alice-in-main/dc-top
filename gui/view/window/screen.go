package window

import "github.com/gdamore/tcell/v2"

var _screen tcell.Screen = nil

func InitScreen(screen tcell.Screen) {
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
