package gui

import (
	"dc-top/docker"

	"github.com/gdamore/tcell/v2"
)

var (
	docker_info_window        Window
	docker_info_data          docker.DockerInfo
	docker_info_border_style  tcell.Style = tcell.StyleDefault.Background(tcell.ColorDarkRed).Foreground(tcell.Color103)
	regular_docker_info_style tcell.Style = tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite)
)

func DockerInfoWindowInit(s tcell.Screen) {
	docker_info_window = NewWindow(s, 106, 1, 130, 30)
}

func DockerInfoWindowDraw() {
	docker_info_window.DrawBorders(docker_info_border_style)
	docker_info_window.DrawContents(dockerInfoDrawerGenerator())
}

func DockerInfoWindowsGet() *Window {
	return &docker_info_window
}

func dockerInfoDrawerGenerator() func(x, y int) (rune, tcell.Style) {
	docker_info_data = docker.GetDockerInfo()
	return func(x, y int) (rune, tcell.Style) {
		return rune(docker_info_data.Info.ID[0]), regular_docker_info_style
	}
}
