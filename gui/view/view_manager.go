package view

import (
	"context"
	"dc-top/docker/compose"
	"dc-top/gui/view/window"
	"dc-top/gui/view/window/bar_window"
	"dc-top/gui/view/window/container_logs_window"
	"dc-top/gui/view/window/containers_window"
	"dc-top/gui/view/window/docker_info_window"
	"dc-top/gui/view/window/edittor_window"
	"dc-top/gui/view/window/error_window"
	"dc-top/gui/view/window/general_info_window"
	"dc-top/gui/view/window/help_window"
	"dc-top/gui/view/window/subshell_window"
	"log"
	"os"
	"sync"

	"github.com/gdamore/tcell/v2"
)

type _viewName uint8

const (
	main _viewName = iota
	main_help
	logs
	logs_help
	edittor
	edittor_help
	subshell
	err
	none
)

var _views map[_viewName]*View = make(map[_viewName]*View)
var _view_stack viewStack = newStack()
var _lock sync.Mutex

func InitDefaultView(bg_context context.Context) {
	_lock.Lock()
	defer _lock.Unlock()

	log.Printf("Initiating default view")
	def_general_info_w := general_info_window.NewGeneralInfoWindow()
	def_containers_w := containers_window.NewContainersWindow()
	def_docker_info_w := docker_info_window.NewDockerInfoWindow()

	controls := help_window.MainControls()
	def_help_w := help_window.NewHelpWindow(
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
	def_bar_w := bar_window.NewBarWindow(bar_dimensions_generator)

	default_view := NewView(map[window.WindowType]window.Window{
		window.GeneralInfo:      &def_general_info_w,
		window.ContainersHolder: &def_containers_w,
		window.DockerInfo:       &def_docker_info_w,
		window.Bar:              &def_bar_w,
		window.Help:             &def_help_w,
	}, window.ContainersHolder,
		tcell.MouseButtonEvents,
		true)

	_views[main] = &default_view
	_view_stack.push(main)
	default_view.Open(bg_context)
}

func ReturnToUpperView() {
	_lock.Lock()
	defer _lock.Unlock()
	if _view_stack.peek() != none {
		log.Println("Returning to previous view")
		_views[currentViewName()].Close()
		_view_stack.pop()
		window.GetScreen().Clear()
		_views[currentViewName()].ResumeWindows()
	} else {
		log.Println("Illegal upper view attempt")
	}
}

func DisplayMainHelp(bg_context context.Context) {
	log.Printf("Changing to main help")
	changeToHelpView(bg_context, main_help, main, help_window.MainControls())
}

func ChangeToLogView(bg_context context.Context, container_id string) {
	log.Printf("Changing to logs")
	logs_window := container_logs_window.NewContainerLogsWindow(container_id)
	bar_dimensions_generator := func() window.Dimensions {
		x1, y1, x2, y2 := window.LogsBarWindowSize()
		return window.NewDimensions(x1, y1, x2, y2, false)
	}
	logs_bar_window := bar_window.NewBarWindow(bar_dimensions_generator)
	logs_view := NewView(map[window.WindowType]window.Window{
		window.ContainerLogs: &logs_window,
		window.Bar:           &logs_bar_window,
	}, window.ContainerLogs,
		0,
		true)
	changeView(bg_context, logs, main, &logs_view)
}

func ChangeToFileEdittor(bg_context context.Context) {
	log.Printf("Changing to edittor")
	file, err := os.Open(compose.DcYamlPath())
	if err != nil {
		log.Printf("Failed to open file %s", compose.DcYamlPath())
		return
	}
	edittor_window := edittor_window.NewEdittorWindow(file)

	edittor_dimensions_generator := func() window.Dimensions {
		x1, y1, x2, y2 := window.LogsBarWindowSize()
		return window.NewDimensions(x1, y1, x2, y2, false)
	}
	edittor_bar_window := bar_window.NewBarWindow(edittor_dimensions_generator)

	edittor_view := NewView(map[window.WindowType]window.Window{
		window.Edittor: &edittor_window,
		window.Bar:     &edittor_bar_window,
	}, window.Edittor,
		0,
		true)
	changeView(bg_context, edittor, main, &edittor_view)
}

func ChangeToSubshell(bg_context context.Context, id string) {
	log.Printf("Changing to subshell")

	subshell_window := subshell_window.NewSubshellWindow(id)

	subshell_view := NewView(map[window.WindowType]window.Window{
		window.Subshell: &subshell_window,
	}, window.Subshell,
		0,
		false)
	changeView(bg_context, subshell, main, &subshell_view)
}

func ChangeToErrorView(bg_context context.Context, message []byte) {
	log.Printf("Changing to error")
	error_window := error_window.NewErrorWindow(message)
	error_view := NewView(map[window.WindowType]window.Window{
		window.Error: &error_window,
	}, window.Error,
		0,
		false)
	changeView(bg_context, err, currentViewName(), &error_view)
}

func DisplayLogHelp(bg_context context.Context) {
	log.Printf("Changing to log help")
	changeToHelpView(bg_context, logs_help, logs, help_window.LogControls())
}

func DisplayEdittorHelp(bg_context context.Context) {
	log.Printf("Changing to edittor help")
	changeToHelpView(bg_context, edittor_help, edittor, help_window.EdittorControls())
}

func HandleKeyPress(key *tcell.EventKey) {
	_lock.Lock()
	defer _lock.Unlock()
	_views[currentViewName()].GetWindow(_views[currentViewName()].GetFocusedWindow()).KeyPress(*key)
}

func HandleMouseEvent(ev *tcell.EventMouse) {
	_lock.Lock()
	defer _lock.Unlock()
	switch currentViewName() {
	case main:
		DefaultView().GetWindow(window.ContainersHolder).MousePress(*ev)
	}
}

func IsCtrlCEnabled() bool {
	return CurrentView().ctrl_c_enabled
}

func CurrentView() *View {
	_lock.Lock()
	defer _lock.Unlock()
	return _views[currentViewName()]
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

func changeToHelpView(bg_context context.Context, new_view_key, prev_view_key _viewName, controls []help_window.Control) {
	help_window := help_window.NewHelpWindow(
		controls,
		func() window.Dimensions {
			width, height := window.GetScreen().Size()
			x1, y1, x2, y2 := width/4, height/4, 3*width/4, (height/4 + len(controls) + 3)
			return window.NewDimensions(x1, y1, x2, y2, true)
		},
	)
	help_view := NewView(map[window.WindowType]window.Window{
		window.Help: &help_window,
	}, window.Help,
		0,
		true)
	changeView(bg_context, new_view_key, prev_view_key, &help_view)
}

func changeView(bg_context context.Context, new_view_key, prev_view_key _viewName, view *View) {
	_lock.Lock()
	defer _lock.Unlock()
	window.GetScreen().Clear()
	_views[prev_view_key].PauseWindows()
	_views[new_view_key] = view
	_view_stack.push(new_view_key)
	view.Open(bg_context)
}

func currentViewName() _viewName {
	return _view_stack.peek()
}
