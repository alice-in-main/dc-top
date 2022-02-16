package gui

import (
	"dc-top/gui/gui_events"
	"dc-top/gui/window"
	"log"
	"os"

	"github.com/gdamore/tcell/v2"
)

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

	windowManager := window.InitWindowManager(s)
	windowManager.OpenAll()

	quit := func() {
		windowManager.CloseAll()
		s.Fini()
		os.Exit(0)
	}
	defer quit()

	for {
		ev := s.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventResize:
			s.Clear()
			windowManager.ResizeAll()
			s.Sync()
		case *tcell.EventKey:
			key := ev.Key()
			switch key {
			case tcell.KeyEscape:
				log.Printf("escaping")
				quit()
			case tcell.KeyCtrlC:
				quit()
			default:
				handleKeyPress(windowManager, ev)
			}
		case *tcell.EventMouse:
			handleMouseEvent(windowManager, ev)
		case gui_events.ChangeToLogsWindowEvent:
			windowManager.SetFocusedWindow(window.ContainerLogs)
			windowManager.CloseAll()
			log.Printf("Changing to logs window of %s", ev.ContainerId)
			s.Clear()
			s.Show()
			new_window := window.NewContainerLogWindow(ev.ContainerId)
			windowManager.Open(window.WindowType(window.ContainerLogs), &new_window)
		default:
			log.Printf("%T", ev)
			log.Printf("GUI got event %s and ignored it\n", ev)
		}
	}
}

func handleKeyPress(wm window.WindowManager, key *tcell.EventKey) {
	switch wm.GetFocusedWindow() {
	case window.Info:
		log.Fatal("shouldnt be here")
	case window.ContainersHolder:
		wm.GetWindow(window.ContainersHolder).KeyPress(*key)
	case window.ContainerLogs:
		wm.GetWindow(window.ContainerLogs).KeyPress(*key)
	}
}

func handleMouseEvent(wm window.WindowManager, ev *tcell.EventMouse) {
	switch wm.GetFocusedWindow() {
	case window.Info:
		log.Fatal("shouldnt be here")
	case window.ContainersHolder:
		wm.GetWindow(window.ContainersHolder).MousePress(*ev)
	case window.ContainerLogs:
		wm.GetWindow(window.ContainerLogs).MousePress(*ev)
	}
}
