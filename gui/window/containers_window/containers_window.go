package containerswindow

import (
	"dc-top/docker"
	"fmt"
	"log"
	"math"
	"strings"

	"github.com/gdamore/tcell/v2"
)

var (
	containers_border_style             = tcell.StyleDefault.Foreground(tcell.Color103)
	regular_container_style tcell.Style = tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite)
	resize_chan                         = make(chan interface{})
	next_index_chan                     = make(chan interface{})
	prev_index_chan                     = make(chan interface{})
	stop_chan                           = make(chan interface{})
	new_container_data_chan             = make(chan docker.ContainerData)
	data_request_chan                   = make(chan tableState)
	draw_queue                          = make(chan tableState)
)

type tableState struct {
	containers_window Window
	index_of_top      int
	table_height      int
	focused_id        string
	containers_data   docker.ContainerData
	sort_type         docker.SortType
}

func ContainersWindowResize() {
	resize_chan <- nil
}

func ContainersWindowNext() {
	next_index_chan <- nil
}

func ContainersWindowPrev() {
	prev_index_chan <- nil
}

func ContainersWindowQuit() {
	stop_chan <- nil
}

func updateIndices(state *tableState, curr_index int) {
	index_of_buttom := state.index_of_top + state.table_height - 1
	if curr_index < state.index_of_top {
		state.index_of_top = curr_index
	} else if curr_index >= index_of_buttom {
		state.index_of_top = curr_index - state.table_height + 1
	}
	log.Printf("CURR: %d, TOP: %d, BUTTOM: %d\n", curr_index, state.index_of_top, index_of_buttom)
}

func drawer() {
	for {
		state := <-draw_queue
		log.Printf("Drawing new state...\n")
		state.containers_window.DrawBorders(containers_border_style)
		state.containers_window.DrawContents(dockerStatsDrawerGenerator(state))
		state.containers_window.screen.Show()
		log.Printf("Done drawing\n")
	}
}

func dockerDataStreamer() {
	for {
		state := <-data_request_chan
		log.Printf("Got request for new data")
		var new_data docker.ContainerData
		if !state.containers_data.AreIdsUpToDate() {
			new_data = docker.GetContainers(&state.containers_data)
		} else {
			state.containers_data.UpdateStats()
			new_data = state.containers_data
		}
		new_data.SortData(state.sort_type)
		log.Printf("Sending back new data")
		new_container_data_chan <- new_data
	}
}

func windowLoop(s tcell.Screen) {
	x1, y1, x2, y2 := containerWindowSize(s)
	state := tableState{
		containers_data:   docker.GetContainers(nil),
		index_of_top:      0,
		table_height:      y2 - y1 - 4 + 1,
		containers_window: NewWindow(s, x1, y1, x2, y2),
		sort_type:         docker.Name,
	}
	log.Printf("table height is %d\n", state.table_height)
	go drawer()
	go dockerDataStreamer()
	log.Printf("Requesting new data\n")
	go func() { data_request_chan <- state }()
	for {
		select {
		case <-resize_chan:
			{
				log.Printf("Resize request\n")
				x1, y1, x2, y2 := containerWindowSize(s)
				state.table_height = y2 - y1 - 4 + 1
				log.Printf("table height is %d\n", state.table_height)
				for i, datum := range state.containers_data.GetData() {
					if datum.ID() == state.focused_id {
						updateIndices(&state, i)
						break
					}
				}
				state.containers_window.Resize(x1, y1, x2, y2)
			}
		case new_data := <-new_container_data_chan:
			{
				log.Printf("Got new data\n")
				state.containers_data = new_data
				log.Printf("Requesting new data\n")
				if !new_data.Contains(state.focused_id) {
					log.Printf("Requesting new data\n")
					state.focused_id = ""
				}
				go func() { data_request_chan <- state }()
			}
		case <-next_index_chan:
			{
				log.Printf("Requesting next index\n")
				if state.focused_id == "" && state.containers_data.Len() > 0 {
					state.focused_id = state.containers_data.GetData()[0].ID()
					log.Printf("Seting first focused_id to %s\n", state.focused_id)
				} else {
					for i, datum := range state.containers_data.GetData() {
						if datum.ID() == state.focused_id {
							var new_index int
							if i == state.containers_data.Len()-1 {
								new_index = 0
							} else {
								new_index = i + 1
							}
							state.focused_id = state.containers_data.GetData()[new_index].ID()
							log.Printf("Changed focused_id to %s\n", state.focused_id)
							updateIndices(&state, new_index)
							break
						}
					}
				}
			}
		case <-prev_index_chan:
			{
				if state.focused_id == "" && state.containers_data.Len() > 0 {
					state.focused_id = state.containers_data.GetData()[state.containers_data.Len()-1].ID()
				} else {
					for i, datum := range state.containers_data.GetData() {
						if datum.ID() == state.focused_id {
							var new_index int
							if i == 0 {
								state.focused_id = state.containers_data.GetData()[state.containers_data.Len()-1].ID()
								new_index = state.containers_data.Len() - 1
							} else {
								state.focused_id = state.containers_data.GetData()[i-1].ID()
								new_index = i - 1
							}
							updateIndices(&state, new_index)
							break
						}
					}
				}
			}
		case <-stop_chan:
			log.Printf("Stopped\n")
			return
		}
		draw_queue <- state
	}
}

func ContainersWindowInit(s tcell.Screen) {
	go windowLoop(s)
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
		num_columns     = len(cell_sizes)
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

func generateDataRow(total_width int, datum *docker.ContainerDatum) (stringStyler, error) {
	stats := datum.CachedStats()
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
			resource_formatter(
				stats.Memory.Usage,
				stats.Memory.Limit,
				"GB"),
			memory_usage_percentage,
			calc_cell_width(memory_cell_percent, total_width),
		),
		PercentageBarDrawer(
			resource_formatter(
				stats.Cpu.ContainerUsage.TotalUsage-stats.PreCpu.ContainerUsage.TotalUsage,
				stats.Cpu.SystemUsage-stats.PreCpu.SystemUsage,
				"GHz"),
			cpu_usage_percentage,
			calc_cell_width(cpu_cell_percent, total_width),
		),
	), nil
}

func generateTable(state *tableState) []stringStyler {
	total_width := (state.containers_window.right_x - 1) - (state.containers_window.left_x + 1)
	underline_rune := '\u2500'
	table := make([]stringStyler, state.containers_data.Len()+2)
	table[0] = generateTableHeader(total_width)
	table[1] = RuneRepeater(underline_rune, tcell.StyleDefault.Foreground(tcell.ColorRebeccaPurple))
	offset := 2
	row_ready_ch := make(chan interface{}, state.containers_data.Len())
	defer close(row_ready_ch)
	for index, datum := range state.containers_data.GetData() {
		go func(i int, d docker.ContainerDatum) {
			row, err := generateDataRow(total_width, &d)
			if err == nil {
				if d.IsDeleted() {
					table[i+offset] = StrikeThrough(row)
				} else {
					table[i+offset] = row
				}
			} else {
				log.Printf("Got error while generating row: %s\n", err)
				table[i+offset] = RuneRepeater(underline_rune, tcell.StyleDefault)
			}
			row_ready_ch <- i
		}(index, datum)
	}
	for range state.containers_data.GetData() {
		<-row_ready_ch
	}
	return table
}

func dockerStatsDrawerGenerator(state tableState) func(x, y int) (rune, tcell.Style) {
	data_table := generateTable(&state)
	log.Printf("New table is ready\n")
	return func(x, y int) (rune, tcell.Style) {
		if y == 0 || y == 1 {
			return data_table[y](x)
		}
		if y+state.index_of_top < len(data_table) {
			r, s := data_table[y+state.index_of_top](x)
			if state.focused_id == state.containers_data.GetData()[y+state.index_of_top-2].ID() {
				s = s.Background(tcell.ColorDarkBlue)
			}
			return r, s
		} else {
			return rune('\x00'), regular_container_style
		}
	}
}
