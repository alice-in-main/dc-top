package view

import (
	"context"
	"dc-top/gui/view/window"
	"dc-top/gui/view/window/bar_window"
	"dc-top/gui/view/window/container_logs_window"
	"dc-top/gui/view/window/containers_window"
	"dc-top/gui/view/window/docker_info_window"
	"dc-top/gui/view/window/general_info_window"
	"dc-top/gui/view/window/help_window"
	"log"
	"sync"

	"github.com/gdamore/tcell/v2"
)

type _viewName uint8

const (
	main _viewName = iota
	logs
	main_help
	logs_help
)

var _views map[_viewName]*View = make(map[_viewName]*View)
var _curr_view _viewName
var _lock sync.Mutex

func InitDefaultView() {
	_lock.Lock()
	defer _lock.Unlock()
	log.Printf("Initiating default view")
	def_general_info_w := general_info_window.NewGeneralInfoWindow(context.Background())
	def_containers_w := containers_window.NewContainersWindow()
	def_docker_info_w := docker_info_window.NewDockerInfoWindow()

	controls := help_window.MainControls()
	def_help_w := help_window.NewHelpWindow(
		context.TODO(),
		controls,
		func() window.Dimensions {
			x1, y1, x2, y2 := window.MainHelpWindowSize()
			return window.NewDimensions(x1, y1, x2, y2, true)
		},
	)

	bar_dimensions_generator := func() window.Dimensions {
		x1, y1, x2, y2 := window.ContainersBarWindowSize()
		return window.NewDimensions(x1, y1, x2, y2, false)
	}
	ctx, cancel := context.WithCancel(context.TODO())
	def_bar_w := bar_window.NewBarWindow(ctx, cancel, bar_dimensions_generator)

	default_view := NewView(map[window.WindowType]window.Window{
		window.GeneralInfo:      &def_general_info_w,
		window.ContainersHolder: &def_containers_w,
		window.DockerInfo:       &def_docker_info_w,
		window.Bar:              &def_bar_w,
		window.Help:             &def_help_w,
	}, window.ContainersHolder)

	_views[main] = &default_view
	_curr_view = main
	_views[_curr_view].Open()
}

func ReturnDefaultView() {
	_lock.Lock()
	defer _lock.Unlock()
	log.Printf("Returning to default view")
	if _curr_view != main {
		log.Printf("Closing %d view and returning to default", _curr_view)
		_views[_curr_view].Close()
		_curr_view = main
		window.GetScreen().Clear()
		_views[main].ResumeWindows()
	} else {
		log.Printf("Tried to run default view when already running default view")
	}
}

func DisplayMainHelp() {
	log.Printf("Changing to main help")
	controls := help_window.MainControls()
	help_window := help_window.NewHelpWindow(
		context.TODO(),
		controls,
		func() window.Dimensions {
			width, height := window.GetScreen().Size()
			x1, y1, x2, y2 := width/3, height/3, 2*width/3, (height/3 + len(controls) + 3)
			return window.NewDimensions(x1, y1, x2, y2, true)
		},
	)
	main_help_view := NewView(map[window.WindowType]window.Window{
		window.Help: &help_window,
	}, window.Help)
	changeView(main_help, main, &main_help_view)
}

func ChangeToLogView(container_id string) {
	log.Printf("Changing to logs")
	logs_window := container_logs_window.NewContainerLogsWindow(container_id)
	bar_dimensions_generator := func() window.Dimensions {
		x1, y1, x2, y2 := window.LogsBarWindowSize()
		return window.NewDimensions(x1, y1, x2, y2, false)
	}
	ctx, cancel := context.WithCancel(context.TODO())
	logs_bar_w := bar_window.NewBarWindow(ctx, cancel, bar_dimensions_generator)
	logs_view := NewView(map[window.WindowType]window.Window{
		window.ContainerLogs: &logs_window,
		window.Bar:           &logs_bar_w,
	}, window.ContainerLogs)
	changeView(logs, main, &logs_view)
}

func HandleKeyPress(key *tcell.EventKey) {
	_lock.Lock()
	defer _lock.Unlock()
	_views[_curr_view].GetWindow(_views[_curr_view].GetFocusedWindow()).KeyPress(*key)
	// switch _curr_view {
	// case main:
	// 	DefaultView().GetWindow(window.ContainersHolder).KeyPress(*key)
	// case logs:
	// 	LogView().GetWindow(window.ContainerLogs).KeyPress(*key)
	// case main_help:
	// 	LogView().GetWindow(window.Help).KeyPress(*key)
	// case logs_help:
	// 	LogView().GetWindow(window.ContainerLogs).KeyPress(*key)
	// }
}

func HandleMouseEvent(ev *tcell.EventMouse) {
	_lock.Lock()
	defer _lock.Unlock()
	switch _curr_view {
	case main:
		DefaultView().GetWindow(window.ContainersHolder).MousePress(*ev)
	}
}

func CurrentView() *View {
	_lock.Lock()
	defer _lock.Unlock()
	return _views[_curr_view]
}

func DefaultView() *View {
	return _views[main]
}

func LogView() *View {
	return _views[logs]
}

func CloseAll() {
	for _, view := range _views {
		view.Close()
	}
}

func changeView(new_view_key, prev_view_key _viewName, view *View) {
	_lock.Lock()
	defer _lock.Unlock()
	_views[prev_view_key].PauseWindows()
	_views[new_view_key] = view
	_curr_view = new_view_key
	view.Open()
}
