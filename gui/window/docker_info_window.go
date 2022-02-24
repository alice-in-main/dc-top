package window

import (
	"dc-top/docker"
	"dc-top/gui/elements"
	"fmt"
	"log"
	"runtime"
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

func (w *DockerInfoWindow) HandleEvent(interface{}) (interface{}, error) {
	log.Println("Info window got event")
	panic(1)
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
		case <-tick.C:
			dockerInfoWindowDraw(state)
		case <-w.stop_chan:
			tick.Stop()
			log.Println("Docker info stopped drawing")
			return
		}
		s.Show()
	}
}

func dockerInfoWindowDraw(state dockerInfoState) {
	DrawBorders(&state.window_state, tcell.StyleDefault)
	DrawContents(&state.window_state, dockerInfoDrawerGenerator())
}

func dockerInfoDrawerGenerator() func(x, y int) (rune, tcell.Style) {
	info_mapper := make(map[int]elements.StringStyler)
	info_arr := make([]elements.StringStyler, 0)
	docker_info_data := docker.GetDockerInfo().Info
	var (
		total   = docker_info_data.Containers
		running = docker_info_data.ContainersRunning
		paused  = docker_info_data.ContainersPaused
		stopped = docker_info_data.ContainersStopped
		col_len = 8
	)
	info_arr = append(info_arr, elements.TextDrawer("Containers summary:", tcell.StyleDefault))
	info_arr = append(info_arr, elements.TextDrawer("total", tcell.StyleDefault).
		Concat(col_len, elements.TextDrawer("running", tcell.StyleDefault)).
		Concat(2*col_len, elements.TextDrawer("paused", tcell.StyleDefault)).
		Concat(3*col_len, elements.TextDrawer("stopped", tcell.StyleDefault)))
	info_arr = append(info_arr, elements.IntegerDrawer(total, tcell.StyleDefault).
		Concat(col_len, elements.IntegerDrawer(running, tcell.StyleDefault)).
		Concat(2*col_len, elements.IntegerDrawer(paused, tcell.StyleDefault)).
		Concat(3*col_len, elements.IntegerDrawer(stopped, tcell.StyleDefault)))
	info_arr = append(info_arr, elements.TextDrawer(fmt.Sprintf("Number of CPUs: %d", docker_info_data.NCPU), tcell.StyleDefault))

	var mem_stats runtime.MemStats
	runtime.ReadMemStats(&mem_stats)
	// total_mem := mem_stats.Sys

	for i, val := range info_arr {
		info_mapper[i] = val
	}
	return func(x, y int) (rune, tcell.Style) {
		if val, ok := info_mapper[y]; ok {
			return val(x)
		} else {
			return ' ', tcell.StyleDefault
		}
	}
}
