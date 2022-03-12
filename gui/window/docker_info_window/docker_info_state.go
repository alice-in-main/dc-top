package docker_info_window

import (
	docker "dc-top/docker"
	"dc-top/gui/elements"
	"dc-top/gui/window"
	"dc-top/gui/window/containers_window"
	"fmt"

	"github.com/gdamore/tcell/v2"
)

type dockerInfoState struct {
	window_state            window.WindowState
	docker_info             docker.DockerInfo
	docker_resource_summary containers_window.TotalStatsSummary
}

func dockerInfoWindowDraw(state dockerInfoState) {
	window.DrawBorders(&state.window_state)
	window.DrawContents(&state.window_state, dockerInfoDrawerGenerator(state))
	window.GetScreen().Show()
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
	window_width := window.Width(&state.window_state)
	max_desc_len := 25
	bar_len := window_width - max_desc_len
	if bar_len < 0 {
		bar_len = 0
	} else if bar_len > 40 {
		bar_len = 40
	}

	docker_cpu_usage := float64(state.docker_resource_summary.TotalCpuUsage)
	system_cpu_usage := float64(state.docker_resource_summary.TotalSystemCpuUsage)
	docker_mem_usage := float64(state.docker_resource_summary.TotalMemUsage)
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
