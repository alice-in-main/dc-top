package gui

import (
	"dc-top/gui/view/window"
	"time"

	"github.com/gdamore/tcell/v2"
)

var (
	upKey         = tcell.NewEventKey(tcell.KeyUp, '\x00', 0)
	downKey       = tcell.NewEventKey(tcell.KeyDown, '\x00', 0)
	inspectKey    = tcell.NewEventKey(tcell.KeyRune, 'i', 0)
	helpKey       = tcell.NewEventKey(tcell.KeyRune, 'h', 0)
	logsKey       = tcell.NewEventKey(tcell.KeyRune, 'l', 0)
	searchKey     = tcell.NewEventKey(tcell.KeyRune, '/', 0)
	clearKey      = tcell.NewEventKey(tcell.KeyRune, 'c', 0)
	enterKey      = tcell.NewEventKey(tcell.KeyEnter, '\x00', 0)
	nextSearchKey = tcell.NewEventKey(tcell.KeyRune, 'n', 0)
	should_pause  = true
)

func tryPause() {
	if should_pause {
		time.Sleep(500 * time.Millisecond)
	}
}

func sendUp() {
	window.GetScreen().PostEvent(upKey)
}

func sendDown() {
	window.GetScreen().PostEvent(downKey)
}

func toggleInspect() {
	window.GetScreen().PostEvent(inspectKey)
}

func toggleHelp() {
	window.GetScreen().PostEvent(helpKey)
}

func toggleLogs() {
	window.GetScreen().PostEvent(logsKey)
}

func startSearch() {
	window.GetScreen().PostEvent(searchKey)
}

func clearSearch() {
	window.GetScreen().PostEvent(clearKey)
}

func enter() {
	window.GetScreen().PostEvent(enterKey)
}

func nextSearchResult() {
	window.GetScreen().PostEvent(nextSearchKey)
}

func typeString(str string) {
	for _, s := range str {
		window.GetScreen().PostEvent(tcell.NewEventKey(tcell.KeyRune, s, 0))
	}
}
