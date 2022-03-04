package gui

import (
	"context"
	"dc-top/gui/window"
	"fmt"
	"log"

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

	finalize := func() {
		windowManager.CloseAll()
		s.Fini()
		if err != nil {
			fmt.Println(err.Error())
		}
	}
	defer finalize()

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
			case tcell.KeyCtrlC:
				return
			default:
				handleKeyPress(windowManager, ev)
			}
		case *tcell.EventMouse:
			handleMouseEvent(windowManager, ev)
		case window.MessageEvent:
			windowManager.GetWindow(ev.Receiver).HandleEvent(ev.Message, ev.Sender)
		case window.ChangeToLogsWindowEvent:
			windowManager.SetFocusedWindow(window.ContainerLogs)
			windowManager.CloseAll()
			log.Printf("Changing to logs window of %s", ev.ContainerId)
			new_window := window.NewContainerLogWindow(ev.ContainerId)
			windowManager.Open(window.WindowType(window.ContainerLogs), &new_window)
		case window.ChangeToLogsShellEvent:
			windowManager.SetFocusedWindow(window.ContainerShell)
			windowManager.CloseAll()
			log.Printf("Changing to shell window of %s", ev.ContainerId)
			new_window := window.NewShellWindow(ev.ContainerId, context.Background())
			windowManager.Open(window.WindowType(window.ContainerShell), &new_window)
		case window.ChangeToViWindowEvent:
			windowManager.SetFocusedWindow(window.Vi)
			windowManager.CloseAll()
			log.Printf("Changing to vi window of %s", ev.FilePath)
			new_window := window.NewViWindow(ev.FilePath, ev.Sender, context.Background())
			windowManager.Open(window.WindowType(window.Vi), &new_window)
		case window.ChangeToDefaultViewEvent:
			log.Printf("Changing back to default")
			windowManager.CloseAll()
			windowManager = window.InitWindowManager(s)
			windowManager.OpenAll()
			windowManager.SetFocusedWindow(window.ContainersHolder)
		case window.FatalErrorEvent:
			err = fmt.Errorf("a fatal error occured at %s:\n%s", ev.When(), ev.Err)
			return
		default:
			log.Printf("%T", ev)
			log.Printf("GUI got event '%s' and ignored it\n", ev)
		}
	}
}

func handleKeyPress(wm window.WindowManager, key *tcell.EventKey) {
	switch wm.GetFocusedWindow() {
	case window.Info:
		log.Fatal("shouldnt be here 1")
	case window.ContainersHolder:
		wm.GetWindow(window.ContainersHolder).KeyPress(*key)
	case window.ContainerLogs:
		wm.GetWindow(window.ContainerLogs).KeyPress(*key)
	}
}

func handleMouseEvent(wm window.WindowManager, ev *tcell.EventMouse) {
	switch wm.GetFocusedWindow() {
	case window.Info:
		log.Fatal("shouldnt be here 2")
	case window.ContainersHolder:
		wm.GetWindow(window.ContainersHolder).MousePress(*ev)
	case window.ContainerLogs:
		wm.GetWindow(window.ContainerLogs).MousePress(*ev)
	}
}
