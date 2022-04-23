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

	mouse_settings tcell.MouseFlags
	ctrl_c_enabled bool
}

func NewView(windows map[window.WindowType]window.Window, focused window.WindowType, mouse_settings tcell.MouseFlags, ctrl_c_enabled bool) View {
	return View{
		windows:        windows,
		focusedWindow:  focused,
		mouse_settings: mouse_settings,
		ctrl_c_enabled: ctrl_c_enabled,
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
	for _, win := range view.windows {
		win.Close()
	}
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
	screen.EnableMouse(view.mouse_settings)
	screen.Sync()
}
