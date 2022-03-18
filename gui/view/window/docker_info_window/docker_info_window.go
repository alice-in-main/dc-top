package docker_info_window

import (
	docker "dc-top/docker"
	"dc-top/gui/view/window"
	"dc-top/gui/view/window/containers_window"
	"log"
	"time"

	"github.com/gdamore/tcell/v2"
)

type DockerInfoWindow struct {
	dimensions     window.Dimensions
	resize_chan    chan interface{}
	new_stats_chan chan containers_window.TotalStatsSummary
	stop_chan      chan interface{}
	enable_toggle  chan bool
}

func NewDockerInfoWindow() DockerInfoWindow {
	x1, y1, x2, y2 := window.DockerInfoWindowSize()
	return DockerInfoWindow{
		dimensions:     window.NewDimensions(x1, y1, x2, y2, true),
		resize_chan:    make(chan interface{}),
		new_stats_chan: make(chan containers_window.TotalStatsSummary),
		stop_chan:      make(chan interface{}),
		enable_toggle:  make(chan bool),
	}
}

func (w *DockerInfoWindow) Open() {
	go w.main()
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

func (w *DockerInfoWindow) HandleEvent(ev interface{}, wt window.WindowType) (interface{}, error) {
	switch ev := ev.(type) {
	case containers_window.TotalStatsSummary:
		w.new_stats_chan <- ev
	default:
		log.Fatalln("Got unknown event in info", ev)
	}
	return nil, nil
}

func (w *DockerInfoWindow) Disable() {
	log.Printf("Disable DockerInfoWindow...")
	w.enable_toggle <- false
}

func (w *DockerInfoWindow) Enable() {
	log.Printf("Enable DockerInfoWindow...")
	w.enable_toggle <- true
}

func (w *DockerInfoWindow) Close() {
	w.stop_chan <- nil
}

func (w *DockerInfoWindow) main() {
	s := window.GetScreen()
	is_enabled := true
	var state dockerInfoState = dockerInfoState{}
	tick := time.NewTicker(1000 * time.Millisecond)
	for {
		select {
		case is_enabled = <-w.enable_toggle:
			log.Printf("changed docker info to %t", is_enabled)
			if is_enabled {
				w.dockerInfoWindowDraw(state)
			}
		case <-w.resize_chan:
			x1, y1, x2, y2 := window.DockerInfoWindowSize()
			w.dimensions.SetBorders(x1, y1, x2, y2)
			if is_enabled {
				w.dockerInfoWindowDraw(state)
			}
		case summary := <-w.new_stats_chan:
			state.docker_resource_summary = summary
			if is_enabled {
				info, err := docker.GetDockerInfo()
				window.ExitIfErr(err)
				state.docker_info = info
				w.dockerInfoWindowDraw(state)
			}
		case <-tick.C:
			var GetTotalStatsRequest = containers_window.GetTotalStats{}
			s.PostEvent(window.NewMessageEvent(window.ContainersHolder, window.DockerInfo, GetTotalStatsRequest))
		case <-w.stop_chan:
			tick.Stop()
			log.Println("Docker info stopped drawing")
			return
		}
	}
}

func (w *DockerInfoWindow) dockerInfoWindowDraw(state dockerInfoState) {
	window.DrawContents(&w.dimensions, dockerInfoDrawerGenerator(state, window.Width(&w.dimensions)))
	window.GetScreen().Show()
}
