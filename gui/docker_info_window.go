package gui

import (
	"dc-top/docker"
	"fmt"

	"github.com/gdamore/tcell/v2"
)

var (
	docker_info_window        Window
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

func dockerInfoDrawerGenerator() func(x, y int) (rune, tcell.Style) {
	info_mapper := make(map[int]stringStyler)
	docker_info_data := docker.GetDockerInfo()
	info_mapper[0] = TextDrawer(docker_info_data.Info.SystemTime, tcell.StyleDefault)
	info_mapper[1] = TextDrawer(fmt.Sprintf("Containers running: %d", docker_info_data.Info.ContainersRunning), tcell.StyleDefault)
	info_mapper[2] = TextDrawer(fmt.Sprintf("Containers paused: %d", docker_info_data.Info.ContainersPaused), tcell.StyleDefault)
	info_mapper[3] = TextDrawer(fmt.Sprintf("Containers stopped: %d", docker_info_data.Info.ContainersStopped), tcell.StyleDefault)
	info_mapper[4] = TextDrawer(fmt.Sprintf("NCPU: %d", docker_info_data.Info.NCPU), tcell.StyleDefault)
	info_mapper[5] = TextDrawer(fmt.Sprintf("MemTotal: %.2fGB", float64(docker_info_data.Info.MemTotal)/float64(1<<30)), tcell.StyleDefault)
	info_mapper[6] = BarDrawer(3.0, 6.4, 6.35, 15)
	info_mapper[7] = BarDrawer(3.0, 6.4, 6.1, 15)
	info_mapper[8] = BarDrawer(3.0, 6.4, 5.3, 15)
	info_mapper[9] = BarDrawer(3.0, 6.4, 3.1, 15)
	info_mapper[9] = BarDrawer(3.0, 6.4, 3.0, 15)
	return func(x, y int) (rune, tcell.Style) {
		if val, ok := info_mapper[y]; ok {
			return val(x)
		} else {
			return ' ', regular_docker_info_style
		}
	}
}
