package docker_info_window

import (
	"context"
	docker "dc-top/docker"
	"dc-top/gui/view/window"
	"dc-top/gui/view/window/containers_window"
	"log"
	"time"

	"github.com/gdamore/tcell/v2"
	"golang.org/x/sync/semaphore"
)

type DockerInfoWindow struct {
	window_ctx       context.Context
	window_cancel    context.CancelFunc
	drawer_semaphore *semaphore.Weighted

	dimensions_generator func() window.Dimensions
	resize_chan          chan interface{}
	new_stats_chan       chan containers_window.TotalStatsSummary
	enable_toggle        chan bool
}

func NewDockerInfoWindow() DockerInfoWindow {
	return DockerInfoWindow{
		drawer_semaphore: semaphore.NewWeighted(1),
		dimensions_generator: func() window.Dimensions {
			x1, y1, x2, y2 := window.DockerInfoWindowSize()
			return window.NewDimensions(x1, y1, x2, y2, true)
		},
		resize_chan:    make(chan interface{}),
		new_stats_chan: make(chan containers_window.TotalStatsSummary),
		enable_toggle:  make(chan bool),
	}
}

func (w *DockerInfoWindow) Open(view_ctx context.Context) {
	w.window_ctx, w.window_cancel = context.WithCancel(view_ctx)
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
	w.drawer_semaphore.Acquire(w.window_ctx, 1)
	w.drawer_semaphore.Release(1)
}

func (w *DockerInfoWindow) Enable() {
	log.Printf("Enable DockerInfoWindow...")
	w.enable_toggle <- true
}

func (w *DockerInfoWindow) Close() {
	w.window_cancel()
}

func (w *DockerInfoWindow) main() {
	s := window.GetScreen()
	is_enabled := true
	var state dockerInfoState = dockerInfoState{}
	tick := time.NewTicker(1000 * time.Millisecond)

	info, err := docker.GetDockerInfo(w.window_ctx)
	window.ExitIfErr(err)
	state.docker_info = info
	w.dockerInfoWindowDraw(state)

	for {
		select {
		case is_enabled = <-w.enable_toggle:
			log.Printf("changed docker info to %t", is_enabled)
			if is_enabled {
				w.dockerInfoWindowDraw(state)
			}
		case <-w.resize_chan:
			if is_enabled {
				w.dockerInfoWindowDraw(state)
			}
		case summary := <-w.new_stats_chan:
			state.docker_resource_summary = summary
			if is_enabled {
				info, err := docker.GetDockerInfo(w.window_ctx)
				window.ExitIfErr(err)
				state.docker_info = info
				w.dockerInfoWindowDraw(state)
			}
		case <-tick.C:
			var GetTotalStatsRequest = containers_window.GetTotalStats{}
			s.PostEvent(window.NewMessageEvent(window.ContainersHolder, window.DockerInfo, GetTotalStatsRequest))
		case <-w.window_ctx.Done():
			tick.Stop()
			log.Println("Docker info stopped drawing")
			return
		}
	}
}

func (w *DockerInfoWindow) dockerInfoWindowDraw(state dockerInfoState) {
	w.drawer_semaphore.Acquire(w.window_ctx, 1)
	dimensions := w.dimensions_generator()
	window.DrawContents(&dimensions, dockerInfoDrawerGenerator(state, window.Width(&dimensions)))
	window.GetScreen().Show()
	w.drawer_semaphore.Release(1)
}
