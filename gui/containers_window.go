package gui

import (
	"dc-top/docker"
	"fmt"
	"log"
	"math"
	"strings"

	"github.com/gdamore/tcell/v2"
)

var (
	containers_window       Window
	focused_container_index int = 0
	containers_data         docker.ContainerData
	docker_info             docker.DockerInfo
	containers_border_style             = tcell.StyleDefault.Background(tcell.ColorDarkRed).Foreground(tcell.Color103)
	regular_container_style tcell.Style = tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite)
	data_table              []stringStyler
)

const (
	num_columns = 5
)

func ContainersWindowInit(s tcell.Screen) {
	containers_window = NewWindow(s, 1, 1, 220, 30)
	containers_data = docker.GetContainers(nil)
	docker_info = docker.GetDockerInfo()
	data_table = generateTable()
}

func ContainersWindowDrawNext() {
	containers_window.DrawBorders(containers_border_style)
	containers_window.DrawContents(dockerStatsDrawerGenerator(true))
}

func ContainersWindowDrawCurr() {
	containers_window.DrawBorders(containers_border_style)
	containers_window.DrawContents(dockerStatsDrawerGenerator(false))
}

func ContainersWindowNext() {
	if focused_container_index < containers_data.Len()-1 {
		focused_container_index++
	} else {
		focused_container_index = 0
	}
}

func ContainersWindowPrev() {
	if focused_container_index > 0 {
		focused_container_index--
	} else {
		focused_container_index = containers_data.Len() - 1
	}
}

func generateTableCell(column_width int, content interface{}) stringStyler {
	switch typed_content := content.(type) {
	case string:
		var cell []rune
		if len(typed_content) < column_width {
			cell = []rune(typed_content + strings.Repeat(" ", column_width-len(typed_content)))
		} else {
			cell = []rune(typed_content[:column_width-3] + "...")
		}
		return func(i int) (rune, tcell.Style) {
			if i >= len(cell) {
				return '\x00', tcell.StyleDefault
			} else {
				return cell[i], tcell.StyleDefault
			}
		}
	case stringStyler:
		return typed_content
	default:
		log.Println("tried to generate table cell from unknown type")
		panic(1)
	}
}

const (
	id_cell_percent     = 0.05
	name_cell_percent   = 0.12
	image_cell_percent  = 0.23
	memory_cell_percent = 0.3
	cpu_cell_percent    = 0.3
)

func generateGenericTableRow(total_width int, cells ...stringStyler) stringStyler {
	const (
		vertical_line_rune = '\u2502'
	)
	var (
		cell_sizes      = []float64{id_cell_percent, name_cell_percent, image_cell_percent, memory_cell_percent, cpu_cell_percent}
		curr_cell_index = 0
		inner_index     = 0
		sum_curr_cells  = 0.0
	)
	return func(i int) (rune, tcell.Style) {
		if i == 0 {
			inner_index = 0
			curr_cell_index = 0
			sum_curr_cells = 0.0
		} else if curr_cell_index < num_columns-1 && float64(i)/float64(total_width) >= sum_curr_cells+cell_sizes[curr_cell_index] {
			sum_curr_cells += cell_sizes[curr_cell_index]
			curr_cell_index++
			inner_index = 0
			return vertical_line_rune, tcell.StyleDefault
		}
		defer func() { inner_index++ }()
		return cells[curr_cell_index](inner_index)
	}
}

func calc_cell_width(relative_size float64, total_width int) int {
	return int(math.Ceil(relative_size * float64(total_width)))
}

func generateTableHeader(total_width int) stringStyler {
	return generateGenericTableRow(
		total_width,
		generateTableCell(calc_cell_width(id_cell_percent, total_width), "ID"),
		generateTableCell(calc_cell_width(name_cell_percent, total_width), "Name"),
		generateTableCell(calc_cell_width(image_cell_percent, total_width), "Image"),
		generateTableCell(calc_cell_width(memory_cell_percent, total_width), "Memory Usage"),
		generateTableCell(calc_cell_width(cpu_cell_percent, total_width), "CPU Usage"),
	)
}

func generateDataRow(total_width int, datum *docker.ContainerDatum) stringStyler {
	//stats := datum.Stats()
	stats := datum.UpdatedStats()
	// log.Println(stats.Name)
	// log.Println(string(datum.Json_stats))
	cpu_usage_percentage := 100.0 * (float64(stats.Cpu.ContainerUsage.TotalUsage) - float64(stats.PreCpu.ContainerUsage.TotalUsage)) / (float64(stats.Cpu.SystemUsage) - float64(stats.PreCpu.SystemUsage))
	memory_usage_percentage := float64(stats.Memory.Usage) / float64(stats.Memory.Limit) * 100.0
	resource_formatter := func(use, limit int64, unit string) string {
		return fmt.Sprintf("%.2f%s/%.2f%s ", float64(use)/float64(1<<30), unit, float64(limit)/float64(1<<30), unit)
	}
	return generateGenericTableRow(
		total_width,
		generateTableCell(calc_cell_width(id_cell_percent, total_width), datum.ID()),
		generateTableCell(calc_cell_width(name_cell_percent, total_width), stats.Name),
		generateTableCell(calc_cell_width(image_cell_percent, total_width), datum.Image()),
		PercentageBarDrawer(
			resource_formatter(stats.Memory.Usage, stats.Memory.Limit, "GB"),
			memory_usage_percentage,
			calc_cell_width(memory_cell_percent, total_width),
		),
		PercentageBarDrawer(
			resource_formatter(stats.Cpu.ContainerUsage.TotalUsage-stats.PreCpu.ContainerUsage.TotalUsage, stats.Cpu.SystemUsage-stats.PreCpu.SystemUsage, "GHz"),
			cpu_usage_percentage,
			calc_cell_width(cpu_cell_percent, total_width),
		),
	)
}

func generateTable() []stringStyler {
	total_width := (containers_window.right_x - 1) - (containers_window.left_x + 1)
	underline_rune := '\u2500'
	table := make([]stringStyler, containers_data.Len()+2)
	table[0] = generateTableHeader(total_width)
	table[1] = RuneRepeater(underline_rune, tcell.StyleDefault.Foreground(tcell.ColorRebeccaPurple))
	offset := 2
	containers_data.SortData(docker.State)
	row_ready_ch := make(chan interface{}, containers_data.Len())
	defer close(row_ready_ch)
	for index, datum := range containers_data.GetData() {
		go func(i int, d docker.ContainerDatum) {
			table[i+offset] = generateDataRow(total_width, &d)
			row_ready_ch <- i
		}(index, datum)
	}
	for range containers_data.GetData() {
		<-row_ready_ch
	}
	return table
}

func dockerStatsDrawerGenerator(is_next bool) func(x, y int) (rune, tcell.Style) {
	if is_next {
		new_docker_info := docker.GetDockerInfo()
		if !docker_info.ContainersEquals(&new_docker_info) {
			docker_info = new_docker_info
			containers_data = docker.GetContainers(&containers_data)
		}
		data_table = generateTable()
	}
	return func(x, y int) (rune, tcell.Style) {
		selected_row := focused_container_index + 2
		if y == selected_row {
			r, s := data_table[y](x)
			return r, s.Background(tcell.ColorDarkBlue)
		} else if y < len(data_table) {
			return data_table[y](x)
		} else {
			return rune('\x00'), regular_container_style
		}
	}
}
