package gui

import (
	"context"
	"dc-top/gui/view"
	"dc-top/gui/view/window"
	"fmt"
	"log"

	"github.com/gdamore/tcell/v2"
)

func Draw() error {
	screen := window.GetScreen()
	screen.EnableMouse(tcell.MouseButtonEvents)

	bg_context, bg_cancel := context.WithCancel(context.Background())

	view.InitDefaultView(bg_context)

	finalize := func() {
		view.CloseAll()
		log.Println("Finished drawing")
	}
	defer finalize()

	for {
		ev := screen.PollEvent()
		switch ev := ev.(type) {
		case *GuiReadinessCheck:
			ev.Ack <- nil
		case *tcell.EventResize:
			screen.Clear()
			view.CurrentView().Resize()
		case *tcell.EventKey:
			key := ev.Key()
			switch key {
			case tcell.KeyCtrlC:
				bg_cancel()
				return nil
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
			// subshells.OpenContainerShell(ev.ContainerId, bg_context)
			view.ChangeToFileEdittor(bg_context)
		case window.ChangeToFileEdittorEvent:
			// subshells.EditDcYaml(bg_context)
			view.ChangeToFileEdittor(bg_context)
		case window.ChangeToLogsWindowEvent:
			view.ChangeToLogView(bg_context, ev.ContainerId)
		case window.ChangeToMainHelpEvent:
			view.DisplayMainHelp(bg_context)
		case window.ChangeToLogsHelpEvent:
			view.DisplayLogHelp(bg_context)
		case window.ChangeToEdittorHelpEvent:
			view.DisplayEdittorHelp(bg_context)
		case window.ReturnUpperViewEvent:
			view.ReturnToUpperView()
		case window.FatalErrorEvent:
			err := fmt.Errorf("a fatal error occured at %s:\n%s", ev.When(), ev.Err)
			log.Println(err.Error())
			bg_cancel()
			return err
		case *tcell.EventError:
			log.Printf("GUI error '%T: %s'\n", ev, ev)
		default:
			log.Printf("GUI got event '%T: %s' and ignored it\n", ev, ev)
		}
	}
}
