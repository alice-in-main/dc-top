package window

type WindowType int8

const (
	ContainersHolder = iota
	Info
)

type WindowManager struct {
	windows       map[WindowType]Window
	focusedWindow WindowType
}

func InitWindowManager() WindowManager {
	containers_w := NewContainersWindow()
	docker_info_w := NewDockerInfoWindow()

	return WindowManager{
		windows: map[WindowType]Window{
			ContainersHolder: &containers_w,
			Info:             &docker_info_w,
		},
		focusedWindow: ContainersHolder,
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
