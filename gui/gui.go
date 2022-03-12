package gui

import (
	"context"
	"dc-top/gui/window"
	"dc-top/gui/window/bar_window"
	"dc-top/gui/window/container_logs_window"
	"dc-top/gui/window/containers_window"
	"dc-top/gui/window/docker_info_window"
	"dc-top/gui/window/general_info_window"
	"dc-top/gui/window/manager"
	"dc-top/gui/window/subshells"
	"fmt"
	"log"

	"github.com/gdamore/tcell/v2"
)

func Draw() {
	screen, err := tcell.NewScreen()
	if err != nil {
		log.Printf("%+v", err)
		panic(1)
	}
	if err = screen.Init(); err != nil {
		log.Printf("%+v", err)
		panic(1)
	}
	window.InitScreen(screen)
	screen.EnableMouse(tcell.MouseButtonEvents)

	windowManager := initWindowManager()
	windowManager.OpenAll()

	finalize := func() {
		windowManager.CloseAll()
		screen.Fini()
		log.Println("Finished drawing")
		if err != nil {
			fmt.Println(err.Error())
		}
	}
	defer finalize()

	for {
		ev := screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventResize:
			screen.Clear()
			windowManager.ResizeAll()
			screen.Sync()
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
		case window.PauseWindowsEvent:
			windowManager.PauseWindows()
		case window.ResumeWindowsEvent:
			windowManager.ResumeWindows()
		case window.ChangeToContainerShellEvent:
			subshells.OpenContainerShell(ev.ContainerId, context.TODO())
		case window.ChangeToFileEdittorEvent:
			subshells.EditDcYaml(context.TODO())
		case window.ChangeToLogsWindowEvent:
			log.Printf("Changing to logs")
			windowManager.PauseWindows()
			logs_window := container_logs_window.NewContainerLogsWindow(ev.ContainerId)
			windowManager.Open(window.ContainerLogs, &logs_window)
			windowManager.SetFocusedWindow(window.ContainerLogs)
		case window.ChangeToDefaultViewEvent:
			log.Printf("Changing back to default")
			windowManager.Close(window.ContainerLogs)
			windowManager.SetFocusedWindow(window.ContainersHolder)
			windowManager.ResumeWindows()
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

func handleKeyPress(wm manager.WindowManager, key *tcell.EventKey) {
	switch wm.GetFocusedWindow() {
	case window.DockerInfo:
		log.Fatal("shouldnt be here 1")
	case window.ContainersHolder:
		wm.GetWindow(window.ContainersHolder).KeyPress(*key)
	case window.ContainerLogs:
		wm.GetWindow(window.ContainerLogs).KeyPress(*key)
	}
}

func handleMouseEvent(wm manager.WindowManager, ev *tcell.EventMouse) {
	switch wm.GetFocusedWindow() {
	case window.DockerInfo:
		log.Fatal("shouldnt be here 2")
	case window.ContainersHolder:
		wm.GetWindow(window.ContainersHolder).MousePress(*ev)
	}
}

func initWindowManager() manager.WindowManager {
	general_info_w := general_info_window.NewGeneralInfoWindow(context.Background())
	containers_w := containers_window.NewContainersWindow()
	docker_info_w := docker_info_window.NewDockerInfoWindow()
	bar_w := bar_window.NewBarWindow(context.Background())

	windows := map[window.WindowType]window.Window{
		window.GeneralInfo:      &general_info_w,
		window.ContainersHolder: &containers_w,
		window.DockerInfo:       &docker_info_w,
		window.Bar:              &bar_w,
	}

	return manager.InitWindowManager(windows, window.ContainersHolder)
}
