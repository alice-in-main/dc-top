package bar_window

import (
	"dc-top/gui/elements"
	"dc-top/gui/view/window"

	"github.com/gdamore/tcell/v2"
)

type BarMessage interface {
	Styler() elements.StringStyler
}

func Info(msg []rune) {
	window.GetScreen().PostEvent(window.NewMessageEvent(window.Bar, window.Other, infoMessage{msg: msg}))
}

func Warn(msg []rune) {
	window.GetScreen().PostEvent(window.NewMessageEvent(window.Bar, window.Other, warnMessage{msg: msg}))
}

func Err(msg []rune) {
	window.GetScreen().PostEvent(window.NewMessageEvent(window.Bar, window.Other, errorMessage{msg: msg}))
}

func Critical(msg []rune) {
	window.GetScreen().PostEvent(window.NewMessageEvent(window.Bar, window.Other, criticalMessage{msg: msg}))
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

type warnMessage struct {
	msg []rune
}

func (m warnMessage) Styler() elements.StringStyler {
	var warn_prefix = []rune("\u26a0 Warn: ")
	return elements.RuneDrawer(warn_prefix, tcell.StyleDefault.Foreground(tcell.ColorYellow)).Concat(len(warn_prefix), elements.RuneDrawer(m.msg, tcell.StyleDefault))
}

type errorMessage struct {
	msg []rune
}

func (m errorMessage) Styler() elements.StringStyler {
	var warn_prefix = []rune("\u26a0 Error: ")
	return elements.RuneDrawer(warn_prefix, tcell.StyleDefault.Foreground(tcell.ColorRed)).Concat(len(warn_prefix), elements.RuneDrawer(m.msg, tcell.StyleDefault))
}

type criticalMessage struct {
	msg []rune
}

func (m criticalMessage) Styler() elements.StringStyler {
	const crit_prefix = "!!!: "
	return elements.TextDrawer(crit_prefix, tcell.StyleDefault.Foreground(tcell.ColorRed)).Concat(len(crit_prefix), elements.RuneDrawer(m.msg, tcell.StyleDefault))
}
