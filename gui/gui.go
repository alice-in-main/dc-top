package gui

import (
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

	quit := func() {
		ContainersWindowQuit()
		s.Fini()
		os.Exit(0)
	}
	defer quit()

	ContainersWindowInit(s)
	DockerInfoWindowInit(s)

	for {
		ev := s.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventResize:
			s.Clear()
			s.Sync()
			ContainersWindowResize()
			DockerInfoWindowResize(s)
		case *tcell.EventKey:
			key := ev.Key()
			switch key {
			case tcell.KeyEscape:
				quit()
			case tcell.KeyCtrlC:
				quit()
			default:
				handleKeyPress(key)
				s.Show()
			}
		default:
			log.Printf("GUI got event %s and ignored it\n", ev)
		}
	}
}

func handleKeyPress(key tcell.Key) {
	switch focusedWindow {
	case windowType(metadata):
		{
			break
		}
	case windowType(containers):
		{
			handleContainersWindowKeyPress(key)
			break
		}
	}
}

func handleContainersWindowKeyPress(key tcell.Key) {
	switch key {
	case tcell.KeyUp:
		ContainersWindowPrev()
	case tcell.KeyDown:
		log.Printf("Asking for next index")
		ContainersWindowNext()
	default:
		return
	}
}
