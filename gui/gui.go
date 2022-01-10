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
		log.Panicf("%+v", err)
	}
	if err := s.Init(); err != nil {
		log.Panicf("%+v", err)
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

	go func() {
		for {
			time.Sleep(1000 * time.Millisecond)
			var tick timedEvent
			s.PostEvent(tick)
		}
	}()
	// TODO split key presses and drawing to different goroutines
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
				ContainersWindowDrawCurr()
				s.Show()
			}
		case timedEvent:
			ContainersWindowDrawNext()
			DockerInfoWindowDraw()
			s.Show()
		}
	}
}

type timedEvent struct{}

func (timedEvent) When() time.Time {
	return time.Now()
}

func handleKeyPress(key tcell.Key) {
	switch focusedWindow {
	case windowType(metadata):
		break
	case windowType(containers):
		handleContainersWindowKeyPress(key)
	}
}

func handleContainersWindowKeyPress(key tcell.Key) {
	switch key {
	case tcell.KeyUp:
		ContainersWindowPrev()
	case tcell.KeyDown:
		ContainersWindowNext()
	}
}
