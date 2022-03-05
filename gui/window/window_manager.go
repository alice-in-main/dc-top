package window

import (
	"context"
	"log"

	"github.com/gdamore/tcell/v2"
)

type WindowType uint8

const (
	ContainersHolder WindowType = iota
	DockerInfo

	Bar
	GeneralInfo
	Other
)

type WindowManager struct {
	windows       map[WindowType]Window
	focusedWindow WindowType
	screen        tcell.Screen
}

func InitWindowManager(screen tcell.Screen) WindowManager {
	general_info_w := NewGeneralInfoWindow(context.Background())
	containers_w := NewContainersWindow()
	docker_info_w := NewDockerInfoWindow()
	bar_w := NewBarWindow(context.Background())

	return WindowManager{
		windows: map[WindowType]Window{
			GeneralInfo:      &general_info_w,
			ContainersHolder: &containers_w,
			DockerInfo:       &docker_info_w,
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

func (wm *WindowManager) CloseAll() {
	for typE, win := range wm.windows {
		log.Printf("Closing %+v", typE)
		win.Close()
	}
	wm.windows = make(map[WindowType]Window)
}

func (wm *WindowManager) PauseWindows() {
	wm.screen.DisableMouse()
	wm.DisableAll()
	wm.screen.Clear()
	wm.screen.Sync()
}

func (wm *WindowManager) ResumeWindows() {
	wm.EnableAll()
	wm.screen.Clear()
	wm.screen.Sync()
	wm.screen.EnableMouse(tcell.MouseButtonEvents)
}

// TODO: move this to other location
func exitIfErr(screen tcell.Screen, err error) {
	if err != nil {
		log.Printf("a fatal error occured: %s\n", err)
		screen.PostEvent(NewFatalErrorEvent(err))
	}
}
