package window

import (
	docker "dc-top/docker"
	"dc-top/gui/elements"
	"fmt"
	"log"
	"time"

	"github.com/gdamore/tcell/v2"
)

type DockerInfoWindow struct {
	resize_chan    chan interface{}
	new_stats_chan chan totalStatsSummary
	stop_chan      chan interface{}
}

type dockerInfoState struct {
	window_state            WindowState
	docker_info             docker.DockerInfo
	docker_resource_summary totalStatsSummary
}

func NewDockerInfoWindow() DockerInfoWindow {
	return DockerInfoWindow{
		resize_chan:    make(chan interface{}),
		new_stats_chan: make(chan totalStatsSummary),
		stop_chan:      make(chan interface{}),
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

func (w *DockerInfoWindow) HandleEvent(ev interface{}, wt WindowType) (interface{}, error) {
	switch ev := ev.(type) {
	case totalStatsSummary:
		w.new_stats_chan <- ev
	default:
		log.Fatalln("Got unknown event in info", ev)
	}
	return nil, nil
}

func (w *DockerInfoWindow) Close() {
	w.stop_chan <- nil
}

func (w *DockerInfoWindow) main(s tcell.Screen) {
	x1, y1, x2, y2 := DockerInfoWindowSize(s)
	var state dockerInfoState = dockerInfoState{
		window_state: NewWindow(x1, y1, x2, y2, NeighboringWindows{UpperNeighbor: true, RightNeighbor: true}),
	}
	tick := time.NewTicker(1000 * time.Millisecond)
	for {
		select {
		case <-w.resize_chan:
			x1, y1, x2, y2 := DockerInfoWindowSize(s)
			state.window_state.SetBorders(x1, y1, x2, y2)
			dockerInfoWindowDraw(s, state)
		case summary := <-w.new_stats_chan:
			state.docker_resource_summary = summary
			state.docker_info = docker.GetDockerInfo()
			dockerInfoWindowDraw(s, state)
		case <-tick.C:
			var getTotalStatsRequest = getTotalStats{}
			s.PostEvent(NewMessageEvent(ContainersHolder, Info, getTotalStatsRequest))
		case <-w.stop_chan:
			tick.Stop()
			log.Println("Docker info stopped drawing")
			return
		}
		s.Show()
	}
}

func dockerInfoWindowDraw(screen tcell.Screen, state dockerInfoState) {
	DrawBorders(screen, &state.window_state)
	DrawContents(screen, &state.window_state, dockerInfoDrawerGenerator(state))
}

func dockerInfoDrawerGenerator(state dockerInfoState) func(x, y int) (rune, tcell.Style) {
	info_mapper := make(map[int]elements.StringStyler)
	info_arr := make([]elements.StringStyler, 0)
	info_arr = append(info_arr, generateTotalContainerStats(&state)...)
	info_arr = append(info_arr, generateResourceUsage(&state)...)
	info_arr = append(info_arr, generateWarnings(&state)...)

	for i, val := range info_arr {
		info_mapper[i] = val
	}
	return func(x, y int) (rune, tcell.Style) {
		if val, ok := info_mapper[y]; ok {
			return val(x)
		} else {
			return '\x00', tcell.StyleDefault
		}
	}
}

func generateTotalContainerStats(state *dockerInfoState) []elements.StringStyler {
	info_arr := make([]elements.StringStyler, 0)
	var (
		total   = state.docker_info.Info.Containers
		running = state.docker_info.Info.ContainersRunning
		paused  = state.docker_info.Info.ContainersPaused
		stopped = state.docker_info.Info.ContainersStopped
		col_len = 8
	)
	info_arr = append(info_arr, elements.TextDrawer("Containers summary:", tcell.StyleDefault.Underline(true)))
	info_arr = append(info_arr, elements.TextDrawer("total", tcell.StyleDefault).
		Concat(col_len, elements.TextDrawer("running", tcell.StyleDefault)).
		Concat(2*col_len, elements.TextDrawer("paused", tcell.StyleDefault)).
		Concat(3*col_len, elements.TextDrawer("stopped", tcell.StyleDefault)))
	info_arr = append(info_arr, elements.IntegerDrawer(total, tcell.StyleDefault).
		Concat(col_len, elements.IntegerDrawer(running, tcell.StyleDefault)).
		Concat(2*col_len, elements.IntegerDrawer(paused, tcell.StyleDefault)).
		Concat(3*col_len, elements.IntegerDrawer(stopped, tcell.StyleDefault)))
	info_arr = append(info_arr, elements.EmptyDrawer())
	return info_arr
}

func generateResourceUsage(state *dockerInfoState) []elements.StringStyler {
	window_width := Width(&state.window_state)
	max_desc_len := 25
	bar_len := window_width - max_desc_len
	if bar_len < 0 {
		bar_len = 0
	} else if bar_len > 40 {
		bar_len = 40
	}

	docker_cpu_usage := float64(state.docker_resource_summary.totalCpuUsage)
	system_cpu_usage := float64(state.docker_resource_summary.totalSystemCpuUsage)
	docker_mem_usage := float64(state.docker_resource_summary.totalMemUsage)
	system_mem_usage := float64(state.docker_info.Info.MemTotal)

	info_arr := make([]elements.StringStyler, 0)
	info_arr = append(info_arr, elements.TextDrawer("Resources summary:", tcell.StyleDefault.Underline(true)))
	info_arr = append(info_arr, elements.TextDrawer(fmt.Sprintf("Number of CPUs: %d", state.docker_info.Info.NCPU), tcell.StyleDefault))
	info_arr = append(info_arr, elements.ValuesBarDrawer("Total CPU usage: ", 0, system_cpu_usage, docker_cpu_usage, bar_len, []rune(fmt.Sprintf(" %.2f%%", 100.0*docker_cpu_usage/system_cpu_usage))))
	info_arr = append(info_arr, elements.ValuesBarDrawer("Total Mem usage: ", 0, system_mem_usage, docker_mem_usage, bar_len, []rune(fmt.Sprintf(" %.2f%%", 100.0*docker_mem_usage/system_mem_usage))))
	info_arr = append(info_arr, elements.EmptyDrawer())
	return info_arr
}

func generateWarnings(state *dockerInfoState) []elements.StringStyler {
	info_arr := make([]elements.StringStyler, 0)
	for _, warning := range state.docker_info.Info.Warnings {
		info_arr = append(info_arr, elements.TextDrawer(warning, tcell.StyleDefault.Foreground(tcell.ColorYellow)))
	}
	info_arr = append(info_arr, elements.EmptyDrawer())
	return info_arr
}
