package window

import (
	"context"
	"log"

	"github.com/gdamore/tcell/v2"
)

type WindowType int8

const (
	ContainersHolder = iota
	Info
	ContainerLogs
	ContainerShell
	Bar
)

type WindowManager struct {
	windows       map[WindowType]Window
	focusedWindow WindowType
	screen        tcell.Screen
}

func InitWindowManager(screen tcell.Screen) WindowManager {
	containers_w := NewContainersWindow()
	docker_info_w := NewDockerInfoWindow()
	bar_w := NewBarWindow(context.Background())

	return WindowManager{
		windows: map[WindowType]Window{
			ContainersHolder: &containers_w,
			Info:             &docker_info_w,
			Bar:              &bar_w,
		},
		focusedWindow: ContainersHolder,
		screen:        screen,
	}
}

func (wm *WindowManager) GetWindow(wt WindowType) Window {
	w := wm.windows[wt]
	return w
}

func (wm *WindowManager) GetFocusedWindow() WindowType {
	return wm.focusedWindow
}

func (wm *WindowManager) SetFocusedWindow(focusedWindow WindowType) {
	wm.focusedWindow = focusedWindow
}

func (wm *WindowManager) OpenAll() {
	for _, win := range wm.windows {
		win.Open(wm.screen)
	}
}

func (wm *WindowManager) Open(t WindowType, new_window Window) {
	if old_window, ok := wm.windows[t]; ok {
		old_window.Close()
	}
	wm.windows[t] = new_window
	new_window.Open(wm.screen)
}

func (wm *WindowManager) ResizeAll() {
	for _, win := range wm.windows {
		win.Resize()
	}
}

func (wm *WindowManager) CloseAll() {
	for _, win := range wm.windows {
		log.Printf("Closing %s", win)
		win.Close()
	}
	wm.windows = make(map[WindowType]Window)
}
