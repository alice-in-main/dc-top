package gui

import (
	"dc-top/gui/window/containers_window"
	"dc-top/gui/window/docker_info_window"
	"log"
	"os"

	"github.com/gdamore/tcell/v2"
)

type windowType int8

const (
	metadata   windowType = 0
	containers windowType = 1
	info       windowType = 2
)

type guiState struct {
	screen             tcell.Screen
	focused_window     windowType
	docker_info_window docker_info_window.DockerInfoWindow
	containers_window  containers_window.ContainersWindow
}

var focusedWindow windowType = windowType(containers)

func Draw() {
	s, err := tcell.NewScreen()
	if err != nil {
		log.Printf("%+v", err)
		panic(1)
	}
	if err = s.Init(); err != nil {
		log.Printf("%+v", err)
		panic(1)
	}
	s.EnableMouse(tcell.MouseButtonEvents)

	state := guiState{
		screen:             s,
		focused_window:     windowType(containers),
		docker_info_window: docker_info_window.NewDockerInfoWindow(),
		containers_window:  containers_window.NewContainersWindow(),
	}

	state.docker_info_window.Open(s)
	state.containers_window.Open(s)

	quit := func() {
		state.docker_info_window.Close()
		state.containers_window.Close()
		s.Fini()
		os.Exit(0)
	}
	defer quit()

	for {
		ev := s.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventResize:
			s.Clear()
			state.containers_window.Resize()
			state.docker_info_window.Resize()
			s.Sync()
		case *tcell.EventKey:
			key := ev.Key()
			switch key {
			case tcell.KeyEscape:
				quit()
			case tcell.KeyCtrlC:
				quit()
			default:
				handleKeyPress(state, key)
			}
		case *tcell.EventMouse:
			handleMouseEvent(state, ev)
		default:
			log.Printf("GUI got event %s and ignored it\n", ev)
		}
	}
}

func handleKeyPress(state guiState, key tcell.Key) {
	switch focusedWindow {
	case windowType(metadata):
		{
			log.Fatal("shouldnt be here")
			break
		}
	case windowType(info):
		{
			log.Fatal("shouldnt be here")
			break
		}
	case windowType(containers):
		{
			state.containers_window.KeyPress(key)
			break
		}
	}
}

func handleMouseEvent(state guiState, ev *tcell.EventMouse) {
	switch focusedWindow {
	case windowType(metadata):
		{
			log.Fatal("shouldnt be here")
			break
		}
	case windowType(info):
		{
			log.Fatal("shouldnt be here")
			break
		}
	case windowType(containers):
		{
			state.containers_window.MousePress(*ev)
			break
		}
	}
}
