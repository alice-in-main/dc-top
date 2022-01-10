package gui

import (
	"dc-top/docker"
	"fmt"

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

var infoMapper map[int]stringStyler = make(map[int]stringStyler)

func dockerInfoDrawerGenerator() func(x, y int) (rune, tcell.Style) {
	docker_info_data = docker.GetDockerInfo()
	infoMapper[0] = TextDrawer(docker_info_data.Info.SystemTime, tcell.StyleDefault)
	infoMapper[1] = TextDrawer(fmt.Sprintf("Containers running: %d", docker_info_data.Info.ContainersRunning), tcell.StyleDefault)
	infoMapper[2] = TextDrawer(fmt.Sprintf("Containers paused: %d", docker_info_data.Info.ContainersPaused), tcell.StyleDefault)
	infoMapper[3] = TextDrawer(fmt.Sprintf("Containers stopped: %d", docker_info_data.Info.ContainersStopped), tcell.StyleDefault)
	infoMapper[4] = TextDrawer(fmt.Sprintf("NCPU: %d", docker_info_data.Info.NCPU), tcell.StyleDefault)
	infoMapper[5] = TextDrawer(fmt.Sprintf("MemTotal: %.2fGB", float64(docker_info_data.Info.MemTotal)/float64(1<<30)), tcell.StyleDefault)
	infoMapper[6] = BarDrawer(3.0, 6.4, 6.35, 15)
	infoMapper[7] = BarDrawer(3.0, 6.4, 6.1, 15)
	infoMapper[8] = BarDrawer(3.0, 6.4, 5.3, 15)
	infoMapper[9] = BarDrawer(3.0, 6.4, 3.1, 15)
	return func(x, y int) (rune, tcell.Style) {
		if val, ok := infoMapper[y]; ok {
			return val(x)
		} else {
			return ' ', regular_docker_info_style
		}
	}
}
