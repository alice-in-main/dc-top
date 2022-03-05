package gui

import (
	"context"
	"dc-top/gui/window"
	"fmt"
	"log"

	"github.com/gdamore/tcell/v2"
)

//TODO: make screen global
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
		log.Println("Finished drawing")
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
		case window.ChangeToLogsEvent:
			container_logs := window.NewContainerLogs(ev.ContainerId)
			container_logs.OpenContainerLogs(s)
		case window.PauseWindowsEvent:
			windowManager.PauseWindows()
		case window.ResumeWindowsEvent:
			windowManager.ResumeWindows()
		case window.ChangeToContainerShellEvent:
			window.OpenContainerShell(ev.ContainerId, context.TODO(), s)
		case window.ChangeToFileEdittorEvent:
			window.EditDcYaml(context.TODO(), s)
		case window.ChangeToDefaultViewEvent:
			log.Printf("Changing back to default")
		case window.FatalErrorEvent:
			err = fmt.Errorf("a fatal error occured at %s:\n%s", ev.When(), ev.Err)
			return
		case *tcell.EventError:
			log.Printf("GUI error '%T: %s'\n", ev, ev)
		default:
			log.Printf("GUI got event '%T: %s' and ignored it\n", ev, ev)
		}
	}
}

func handleKeyPress(wm window.WindowManager, key *tcell.EventKey) {
	switch wm.GetFocusedWindow() {
	case window.DockerInfo:
		log.Fatal("shouldnt be here 1")
	case window.ContainersHolder:
		wm.GetWindow(window.ContainersHolder).KeyPress(*key)
	}
}

func handleMouseEvent(wm window.WindowManager, ev *tcell.EventMouse) {
	switch wm.GetFocusedWindow() {
	case window.DockerInfo:
		log.Fatal("shouldnt be here 2")
	case window.ContainersHolder:
		wm.GetWindow(window.ContainersHolder).MousePress(*ev)
	}
}
