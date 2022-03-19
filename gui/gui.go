package gui

import (
	"context"
	"dc-top/gui/view"
	"dc-top/gui/view/window"
	"dc-top/gui/view/window/subshells"
	"log"

	"github.com/gdamore/tcell/v2"
)

func Draw() {
	screen := window.GetScreen()
	screen.EnableMouse(tcell.MouseButtonEvents)

	view.InitDefaultView()

	finalize := func() {
		view.CloseAll()
		log.Println("Finished drawing")
	}
	defer finalize()

	for {
		ev := screen.PollEvent()
		switch ev := ev.(type) {
		case *ReadinessCheck:
			ev.Ack <- nil
		case *tcell.EventResize:
			screen.Clear()
			view.CurrentView().Resize()
			screen.Sync()
		case *tcell.EventKey:
			key := ev.Key()
			switch key {
			case tcell.KeyCtrlC:
				return
			default:
				view.HandleKeyPress(ev)
			}
		case *tcell.EventMouse:
			view.HandleMouseEvent(ev)
		case window.MessageEvent:
			if view.CurrentView().Exists(ev.Receiver) {
				view.CurrentView().GetWindow(ev.Receiver).HandleEvent(ev.Message, ev.Sender)
			}
		case window.PauseWindowsEvent:
			view.CurrentView().PauseWindows()
		case window.ResumeWindowsEvent:
			view.CurrentView().ResumeWindows()
		case window.ChangeToContainerShellEvent:
			subshells.OpenContainerShell(ev.ContainerId, context.TODO())
		case window.ChangeToFileEdittorEvent:
			subshells.EditDcYaml(context.TODO())
		case window.ChangeToLogsWindowEvent:
			view.ChangeToLogView(ev.ContainerId)
		case window.ChangeToMainHelpEvent:
			view.DisplayMainHelp()
		case window.ChangeToLogsHelpEvent:
			view.DisplayLogHelp()
		case window.ReturnUpperViewEvent:
			view.ReturnToUpperView()
		case window.FatalErrorEvent:
			log.Printf("a fatal error occured at %s:\n%s", ev.When(), ev.Err)
			return
		case *tcell.EventError:
			log.Printf("GUI error '%T: %s'\n", ev, ev)
		default:
			log.Printf("GUI got event '%T: %s' and ignored it\n", ev, ev)
		}
	}
}
