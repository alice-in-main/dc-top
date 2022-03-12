package manager

import (
	"dc-top/gui/window"
	"log"

	"github.com/gdamore/tcell/v2"
)

type WindowManager struct {
	windows       map[window.WindowType]window.Window
	focusedWindow window.WindowType
}

func InitWindowManager(windows map[window.WindowType]window.Window, focused window.WindowType) WindowManager {
	return WindowManager{
		windows:       windows,
		focusedWindow: focused,
	}
}

func (wm *WindowManager) GetWindow(wt window.WindowType) window.Window {
	w := wm.windows[wt]
	return w
}

func (wm *WindowManager) GetFocusedWindow() window.WindowType {
	return wm.focusedWindow
}

func (wm *WindowManager) SetFocusedWindow(focusedWindow window.WindowType) {
	wm.focusedWindow = focusedWindow
}

func (wm *WindowManager) OpenAll() {
	for _, win := range wm.windows {
		win.Open()
	}
}

func (wm *WindowManager) Open(t window.WindowType, new_window window.Window) {
	if old_window, ok := wm.windows[t]; ok {
		old_window.Close()
	}
	wm.windows[t] = new_window
	new_window.Open()
}

func (wm *WindowManager) ResizeAll() {
	for _, win := range wm.windows {
		win.Resize()
	}
}

func (wm *WindowManager) EnableAll() {
	for _, win := range wm.windows {
		win.Enable()
	}
}

func (wm *WindowManager) DisableAll() {
	for _, win := range wm.windows {
		win.Disable()
	}
}

func (wm *WindowManager) Close(t window.WindowType) {
	wm.GetWindow(t).Close()
	delete(wm.windows, t)
}

func (wm *WindowManager) CloseAll() {
	for typE, win := range wm.windows {
		log.Printf("Closing %+v", typE)
		win.Close()
	}
	wm.windows = make(map[window.WindowType]window.Window)
}

func (wm *WindowManager) PauseWindows() {
	screen := window.GetScreen()
	screen.DisableMouse()
	wm.DisableAll()
	screen.Clear()
	screen.Sync()
}

func (wm *WindowManager) ResumeWindows() {
	screen := window.GetScreen()
	wm.EnableAll()
	screen.Clear()
	screen.Sync()
	screen.EnableMouse(tcell.MouseButtonEvents)
}
