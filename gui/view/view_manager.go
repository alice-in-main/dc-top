package view

import (
	"context"
	"dc-top/gui/view/window"
	"dc-top/gui/view/window/bar_window"
	"dc-top/gui/view/window/container_logs_window"
	"dc-top/gui/view/window/containers_window"
	"dc-top/gui/view/window/docker_info_window"
	"dc-top/gui/view/window/general_info_window"
	"log"
	"sync"

	"github.com/gdamore/tcell/v2"
)

type _viewName uint8

const (
	def _viewName = iota
	logs
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

	bar_dimensions_generator := func() window.Dimensions {
		x1, y1, x2, y2 := window.ContainersBarWindowSize()
		return window.NewDimensions(x1, y1, x2, y2, false)
	}
	def_bar_w := bar_window.NewBarWindow(context.Background(), bar_dimensions_generator)

	default_view := NewView(map[window.WindowType]window.Window{
		window.GeneralInfo:      &def_general_info_w,
		window.ContainersHolder: &def_containers_w,
		window.DockerInfo:       &def_docker_info_w,
		window.Bar:              &def_bar_w,
	}, window.ContainersHolder)

	_views[def] = &default_view
	_curr_view = def
	_views[_curr_view].Open()
}

func ChangeToLogView(container_id string) {
	_lock.Lock()
	defer _lock.Unlock()
	log.Printf("Changing to logs")
	_views[def].PauseWindows()
	logs_window := container_logs_window.NewContainerLogsWindow(container_id)

	bar_dimensions_generator := func() window.Dimensions {
		x1, y1, x2, y2 := window.LogsBarWindowSize()
		return window.NewDimensions(x1, y1, x2, y2, false)
	}
	logs_bar_w := bar_window.NewBarWindow(context.Background(), bar_dimensions_generator)
	logs_view := NewView(map[window.WindowType]window.Window{
		window.ContainerLogs: &logs_window,
		window.Bar:           &logs_bar_w,
	}, window.ContainerLogs)
	_views[logs] = &logs_view
	_curr_view = logs
	logs_view.Open()
}

func RunDefaultView() {
	_lock.Lock()
	defer _lock.Unlock()
	log.Printf("Returning to default view")
	if _curr_view != def {
		log.Printf("Closing %d view and returning to default", _curr_view)
		_views[_curr_view].Close()
		_curr_view = def
		_views[def].ResumeWindows()
	} else {
		log.Printf("Tried to run default view when already running default view")
	}
}

func HandleKeyPress(key *tcell.EventKey) {
	_lock.Lock()
	defer _lock.Unlock()
	switch _curr_view {
	case def:
		DefaultView().GetWindow(window.ContainersHolder).KeyPress(*key)
	case logs:
		LogView().GetWindow(window.ContainerLogs).KeyPress(*key)
	}
}

func HandleMouseEvent(ev *tcell.EventMouse) {
	_lock.Lock()
	defer _lock.Unlock()
	switch _curr_view {
	case def:
		DefaultView().GetWindow(window.ContainersHolder).MousePress(*ev)
	}
}

func CurrentView() *View {
	_lock.Lock()
	defer _lock.Unlock()
	return _views[_curr_view]
}

func DefaultView() *View {
	return _views[def]
}

func LogView() *View {
	return _views[logs]
}

func CloseAll() {
	for _, view := range _views {
		view.Close()
	}
}
