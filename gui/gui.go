package gui

import (
	"log"
	"os"
	"time"

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
	if err := s.Init(); err != nil {
		log.Printf("%+v", err)
		panic(1)
	}

	quit := func() {
		s.Fini()
		os.Exit(0)
	}
	defer quit()

	ContainersWindowInit(s)
	ContainersWindowDrawNext()

	DockerInfoWindowInit(s)
	DockerInfoWindowDraw()

	go startTicking(s)
	for {
		ev := s.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventResize:
			s.Sync()
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
			ContainersWindowDrawNext()
			DockerInfoWindowDraw()
			s.Show()
		}
	}
}

func startTicking(screen tcell.Screen) {
	for {
		time.Sleep(1000 * time.Millisecond)
		var tick tickEvent
		screen.PostEvent(tick)
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
		ContainersWindowNext()
	}
	ContainersWindowDrawCurr()
}
