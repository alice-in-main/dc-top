package gui

import (
	"dc-top/docker"

	"github.com/gdamore/tcell/v2"
)

var (
	containers_window       Window
	focused_container_index int = 0
	containers_data         []docker.ContainerData
	parsed_containers_data  []string
	containers_border_style             = tcell.StyleDefault.Background(tcell.ColorDarkRed).Foreground(tcell.Color103)
	regular_container_style tcell.Style = tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite)
	focused_container_style tcell.Style = tcell.StyleDefault.Background(tcell.ColorWhite).Foreground(tcell.ColorBlack)
)

func ContainersWindowInit(s tcell.Screen) {
	containers_window = NewWindow(s, 1, 1, 104, 30)
	containers_data = docker.GetContainers()
	parsed_containers_data = make([]string, len(containers_data))
	for i := 0; i < len(containers_data); i++ {
		parsed_containers_data[i] = containers_data[i].String()
	}
}

func ContainersWindowDrawNext() {
	containers_window.DrawBorders(containers_border_style)
	containers_window.DrawContents(dockerStatsDrawerGenerator(true))
	fetchNewContainerData()
}

func ContainersWindowDrawCurr() {
	containers_window.DrawBorders(containers_border_style)
	containers_window.DrawContents(dockerStatsDrawerGenerator(false))
}

func ContainersWindowNext() {
	if focused_container_index < len(containers_data)-1 {
		focused_container_index++
	} else {
		focused_container_index = 0
	}
}

func ContainersWindowPrev() {
	if focused_container_index > 0 {
		focused_container_index--
	} else {
		focused_container_index = len(containers_data) - 1
	}
}

func fetchNewContainerData() {
	for _, datum := range containers_data {
		datum.Close()
	}
	containers_data = docker.GetContainers()
}

func dockerStatsDrawerGenerator(is_next bool) func(x, y int) (rune, tcell.Style) {
	if is_next {
		parsed_containers_data = make([]string, len(containers_data))
		for i := 0; i < len(containers_data); i++ {
			parsed_containers_data[i] = containers_data[i].String()
		}
	}
	return func(x, y int) (rune, tcell.Style) {
		if y == focused_container_index && x < len(parsed_containers_data[y]) {
			return rune(parsed_containers_data[y][x]), focused_container_style
		} else if y < len(parsed_containers_data) && x < len(parsed_containers_data[y]) {
			return rune(parsed_containers_data[y][x]), regular_container_style
		} else {
			return rune('\x00'), regular_container_style
		}
	}
}
