package window

import (
	"dc-top/docker"
	"dc-top/gui/elements"
	"fmt"
	"log"
	"time"

	"github.com/gdamore/tcell/v2"
)

type DockerInfoWindow struct {
	resize_chan chan interface{}
	stop_chan   chan interface{}
}

type dockerInfoState struct {
	window_state WindowState
}

func NewDockerInfoWindow() DockerInfoWindow {
	return DockerInfoWindow{
		resize_chan: make(chan interface{}),
		stop_chan:   make(chan interface{}),
	}
}

func (w *DockerInfoWindow) Open(s tcell.Screen) {
	go w.main(s)
}

func (w *DockerInfoWindow) Resize() {
	w.resize_chan <- nil
}

func (w *DockerInfoWindow) KeyPress(key tcell.EventKey) {
	log.Fatal("docker info key press isn't implemented")
}

func (w *DockerInfoWindow) MousePress(tcell.EventMouse) {
	log.Fatal("docker info mouse press isn't implemented")
}

func (w *DockerInfoWindow) Close() {
	w.stop_chan <- nil
}

func (w *DockerInfoWindow) main(s tcell.Screen) {
	x1, y1, x2, y2 := DockerInfoWindowSize(s)
	var state dockerInfoState = dockerInfoState{
		window_state: NewWindow(s, x1, y1, x2, y2),
	}
	tick := time.NewTicker(1000 * time.Millisecond)
	for {
		select {
		case <-w.resize_chan:
			x1, y1, x2, y2 := DockerInfoWindowSize(s)
			state.window_state.SetBorders(x1, y1, x2, y2)
			dockerInfoWindowDraw(state)
		case <-w.stop_chan:
			tick.Stop()
			log.Println("Docker info stopped drawing")
			return
		case <-tick.C:
			dockerInfoWindowDraw(state)
		}
		s.Show()
	}
}

func dockerInfoWindowDraw(state dockerInfoState) {
	DrawBorders(&state.window_state, tcell.StyleDefault.Background(tcell.ColorDarkRed).Foreground(tcell.Color103))
	DrawContents(&state.window_state, dockerInfoDrawerGenerator())
}

func dockerInfoDrawerGenerator() func(x, y int) (rune, tcell.Style) {
	info_mapper := make(map[int]elements.StringStyler)
	docker_info_data := docker.GetDockerInfo()
	info_mapper[0] = elements.TextDrawer(docker_info_data.Info.SystemTime, tcell.StyleDefault)
	info_mapper[1] = elements.TextDrawer(fmt.Sprintf("Containers running: %d", docker_info_data.Info.ContainersRunning), tcell.StyleDefault)
	info_mapper[2] = elements.TextDrawer(fmt.Sprintf("Containers paused: %d", docker_info_data.Info.ContainersPaused), tcell.StyleDefault)
	info_mapper[3] = elements.TextDrawer(fmt.Sprintf("Containers stopped: %d", docker_info_data.Info.ContainersStopped), tcell.StyleDefault)
	info_mapper[4] = elements.TextDrawer(fmt.Sprintf("NCPU: %d", docker_info_data.Info.NCPU), tcell.StyleDefault)
	info_mapper[5] = elements.TextDrawer(fmt.Sprintf("NCPU: %d", docker_info_data.Info.NCPU), tcell.StyleDefault)
	info_mapper[6] = elements.TextDrawer(fmt.Sprintf("MemTotal: %.2fGB", float64(docker_info_data.Info.MemTotal)/float64(1<<30)), tcell.StyleDefault)
	info_mapper[7] = elements.ValuesBarDrawer("", 3.0, 6.4, 6.35, 15)
	info_mapper[8] = elements.ValuesBarDrawer("", 3.0, 6.4, 6.1, 15)
	info_mapper[9] = elements.ValuesBarDrawer("", 3.0, 6.4, 5.3, 15)
	info_mapper[10] = elements.ValuesBarDrawer("", 3.0, 6.4, 3.1, 15)
	info_mapper[11] = elements.ValuesBarDrawer("", 3.0, 6.4, 3.0, 15)
	info_mapper[12] = elements.ValuesBarDrawer("", 3.0, 6.4, 3.5, 15)
	info_mapper[13] = elements.PercentageBarDrawer("", 0.0, 15)
	info_mapper[14] = elements.PercentageBarDrawer("", 0.1, 15)
	info_mapper[15] = elements.PercentageBarDrawer("", 3.0, 15)
	info_mapper[16] = elements.PercentageBarDrawer("", 16, 15)
	info_mapper[17] = elements.PercentageBarDrawer("", 40, 15)
	info_mapper[18] = elements.PercentageBarDrawer("", 70, 15)

	return func(x, y int) (rune, tcell.Style) {
		if val, ok := info_mapper[y]; ok {
			return val(x)
		} else {
			return ' ', tcell.StyleDefault
		}
	}
}
