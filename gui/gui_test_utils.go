package gui

import (
	"dc-top/gui/view/window"
	"time"

	"github.com/gdamore/tcell/v2"
)

var (
	upKey          = tcell.NewEventKey(tcell.KeyUp, '\x00', 0)
	downKey        = tcell.NewEventKey(tcell.KeyDown, '\x00', 0)
	eofKey         = tcell.NewEventKey(tcell.KeyCtrlD, '\x00', 0)
	inspectKey     = tcell.NewEventKey(tcell.KeyRune, 'i', 0)
	helpKey        = tcell.NewEventKey(tcell.KeyRune, 'h', 0)
	logsKey        = tcell.NewEventKey(tcell.KeyRune, 'l', 0)
	searchKey      = tcell.NewEventKey(tcell.KeyRune, '/', 0)
	clearKey       = tcell.NewEventKey(tcell.KeyRune, 'c', 0)
	enterKey       = tcell.NewEventKey(tcell.KeyEnter, '\x00', 0)
	nextSearchKey  = tcell.NewEventKey(tcell.KeyRune, 'n', 0)
	subshellKey    = tcell.NewEventKey(tcell.KeyRune, 'e', 0)
	edittorKey     = tcell.NewEventKey(tcell.KeyRune, 'v', 0)
	saveEdittorKey = tcell.NewEventKey(tcell.KeyCtrlS, '\x00', 0)
	quitEdittorKey = tcell.NewEventKey(tcell.KeyCtrlQ, '\x00', 0)
	lineDeleteKey  = tcell.NewEventKey(tcell.KeyCtrlD, '\x00', 0)
	should_pause   = true
)

func _post_event_with_delay(ev *tcell.EventKey) {
	window.GetScreen().PostEvent(ev)
	time.Sleep(20 * time.Millisecond)
}

func tryPause() {
	if should_pause {
		time.Sleep(500 * time.Millisecond)
	}
}

func sendUp() {
	_post_event_with_delay(upKey)
}

func sendDown() {
	_post_event_with_delay(downKey)
}

func toggleInspect() {
	_post_event_with_delay(inspectKey)
}

func toggleHelp() {
	_post_event_with_delay(helpKey)
}

func toggleLogs() {
	_post_event_with_delay(logsKey)
}

func enterSubshell() {
	_post_event_with_delay(subshellKey)
}

func startSearch() {
	_post_event_with_delay(searchKey)
}

func clearSearch() {
	_post_event_with_delay(clearKey)
}

func enter() {
	_post_event_with_delay(enterKey)
}

func nextSearchResult() {
	_post_event_with_delay(nextSearchKey)
}

func typeString(str string) {
	for _, s := range str {
		_post_event_with_delay(tcell.NewEventKey(tcell.KeyRune, s, 0))
	}
}

func sendEof() {
	_post_event_with_delay(eofKey)
}

func enterEdittor() {
	_post_event_with_delay(edittorKey)
}

func saveEdittor() {
	_post_event_with_delay(saveEdittorKey)
}

func deleteLineInEdittor() {
	_post_event_with_delay(lineDeleteKey)
}

func quitEdittorWithoutSaving() {
	_post_event_with_delay(quitEdittorKey)
}
