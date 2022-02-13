package gui

import (
	"dc-top/gui/window"
	"log"
	"os"

	"github.com/gdamore/tcell/v2"
)

type guiState struct {
	screen        tcell.Screen
	windowManager window.WindowManager
}

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
		screen:        s,
		windowManager: window.InitWindowManager(),
	}

	state.windowManager.GetWindow(window.ContainersHolder).Open(s)
	state.windowManager.GetWindow(window.Info).Open(s)

	quit := func() {
		state.windowManager.GetWindow(window.ContainersHolder).Close()
		state.windowManager.GetWindow(window.Info).Close()
		s.Fini()
		os.Exit(0)
	}
	defer quit()

	for {
		ev := s.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventResize:
			s.Clear()
			state.windowManager.GetWindow(window.ContainersHolder).Resize()
			state.windowManager.GetWindow(window.Info).Resize()
			s.Sync()
		case *tcell.EventKey:
			key := ev.Key()
			switch key {
			case tcell.KeyEscape:
				log.Printf("Escaping")
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
	switch state.windowManager.GetFocusedWindow() {
	case window.WindowType(window.Info):
		{
			log.Fatal("shouldnt be here")
			break
		}
	case window.WindowType(window.ContainersHolder):
		{
			state.windowManager.GetWindow(window.ContainersHolder).KeyPress(key)
			break
		}
	}
}

func handleMouseEvent(state guiState, ev *tcell.EventMouse) {
	switch state.windowManager.GetFocusedWindow() {
	case window.WindowType(window.Info):
		{
			log.Fatal("shouldnt be here")
			break
		}
	case window.WindowType(window.ContainersHolder):
		{
			state.windowManager.GetWindow(window.ContainersHolder).MousePress(*ev)
			break
		}
	}
}
