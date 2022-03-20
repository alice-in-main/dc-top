package view

import (
	"context"
	"dc-top/gui/view/window"

	"github.com/gdamore/tcell/v2"
)

type View struct {
	view_ctx    context.Context
	view_cancel context.CancelFunc

	windows       map[window.WindowType]window.Window
	focusedWindow window.WindowType
}

func NewView(windows map[window.WindowType]window.Window, focused window.WindowType) View {
	return View{
		windows:       windows,
		focusedWindow: focused,
	}
}

func (view *View) Exists(wt window.WindowType) bool {
	_, ok := view.windows[wt]
	return ok
}

func (view *View) GetWindow(wt window.WindowType) window.Window {
	w := view.windows[wt]
	return w
}

func (view *View) GetFocusedWindow() window.WindowType {
	return view.focusedWindow
}

func (view *View) SetFocusedWindow(focusedWindow window.WindowType) {
	view.focusedWindow = focusedWindow
}

func (view *View) Open(bg_context context.Context) {
	view.view_ctx, view.view_cancel = context.WithCancel(bg_context)

	for _, win := range view.windows {
		win.Open(view.view_ctx)
	}
}

func (view *View) Resize() {
	for _, win := range view.windows {
		win.Resize()
	}
}

func (view *View) Enable() {
	for _, win := range view.windows {
		win.Enable()
	}
}

func (view *View) Disable() {
	for _, win := range view.windows {
		win.Disable()
	}
}

func (view *View) Close() {
	view.view_cancel()
	view.windows = make(map[window.WindowType]window.Window)
}

func (view *View) PauseWindows() {
	screen := window.GetScreen()
	screen.DisableMouse()
	view.Disable()
}

func (view *View) ResumeWindows() {
	screen := window.GetScreen()
	view.Enable()
	screen.EnableMouse(tcell.MouseButtonEvents)
	screen.Sync()
}