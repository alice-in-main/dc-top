package containers_window

import (
	docker "dc-top/docker"
	"dc-top/gui/elements"
	"dc-top/gui/window"
	"fmt"
	"log"
	"math"
	"strings"

	"github.com/gdamore/tcell/v2"
)

const (
	id_cell_percent     = 0.04
	state_cell_percent  = 0.04
	name_cell_percent   = 0.12
	image_cell_percent  = 0.24
	memory_cell_percent = 0.28
	cpu_cell_percent    = 0.28
)

func dockerStatsDrawerGenerator(state tableState) (func(x, y int) (rune, tcell.Style), error) {
	if state.window_mode == containers {
		data_table := generateTable(&state)
		search_row := state.search_box.Style()
		search_filter_message := elements.TextDrawer(fmt.Sprintf("Showing only containers containing '%s'", state.search_box.Value()), tcell.StyleDefault)

		return func(x, y int) (rune, tcell.Style) {
			if y == 0 || y == 1 {
				return data_table[y](x)
			}
			if y == state.table_height+2 {
				if state.keyboard_mode == search {
					return search_row(x)
				} else if state.keyboard_mode == regular && state.search_box.Value() != "" {
					return search_filter_message(x)
				}
			}
			if y+state.index_of_top_container < len(data_table) {
				r, s := data_table[y+state.index_of_top_container](x)
				if state.focused_id == state.filtered_data[y+state.index_of_top_container-2].ID() {
					s = s.Background(tcell.ColorDarkBlue)
				}
				return r, s
			} else {
				return rune('\x00'), tcell.StyleDefault
			}
		}, nil
	} else if state.window_mode == inspect {
		pretty_info, err := generatePrettyInspectInfo(state)
		if err != nil {
			return func(_, _ int) (rune, tcell.Style) { return '\x00', tcell.StyleDefault }, err
		}
		return func(x, y int) (rune, tcell.Style) {
			if val, ok := pretty_info[y]; ok {
				return (val)(x)
			} else {
				return '\x00', tcell.StyleDefault
			}
		}, nil
	} else {
		log.Printf("Got into unimplemented containers window mode '%d'", state.window_mode)
		panic(1)
	}
}

func getCellWidths(total_width int) []int {
	return []int{
		int(id_cell_percent * float64(total_width)),
		int(state_cell_percent * float64(total_width)),
		int(name_cell_percent * float64(total_width)),
		int(image_cell_percent * float64(total_width)),
		int(memory_cell_percent * float64(total_width)),
		int(cpu_cell_percent * float64(total_width)),
	}
}

func generateTableCell(column_width int, content interface{}) elements.StringStyler {
	switch typed_content := content.(type) {
	case string:
		var cell []rune
		if len(typed_content) < column_width {
			cell = []rune(typed_content + strings.Repeat(" ", column_width-len(typed_content)))
		} else {
			num_dots := (column_width - 1) / 3
			if num_dots > 3 {
				num_dots = 3
			}
			cell = []rune(typed_content[:column_width-num_dots] + strings.Repeat(".", num_dots))
		}
		return func(i int) (rune, tcell.Style) {
			if i >= len(cell) {
				return '\x00', tcell.StyleDefault
			} else {
				return cell[i], tcell.StyleDefault
			}
		}
	case elements.StringStyler:
		return typed_content
	default:
		log.Println("tried to generate table cell from unknown type")
		panic(1)
	}
}

func generateGenericTableRow(total_width int, cells ...elements.StringStyler) elements.StringStyler {
	const (
		vertical_line_rune = '\u2502'
	)
	var (
		cell_sizes      = getCellWidths(total_width)
		num_columns     = len(cell_sizes)
		curr_cell_index = 0
		inner_index     = 0
	)

	return func(i int) (rune, tcell.Style) {
		if i == 0 {
			inner_index = 0
			curr_cell_index = 0
		} else if curr_cell_index < num_columns-1 && inner_index == cell_sizes[curr_cell_index] {
			curr_cell_index++
			inner_index = 0
			return vertical_line_rune, tcell.StyleDefault
		}
		defer func() { inner_index++ }()
		return cells[curr_cell_index](inner_index)
	}
}

func calcCellWidth(relative_size float64, total_width int) int {
	return int(math.Ceil(relative_size * float64(total_width)))
}

func generateTableHeader(total_width int) elements.StringStyler {
	return generateGenericTableRow(
		total_width,
		generateTableCell(calcCellWidth(id_cell_percent, total_width), "ID"),
		generateTableCell(calcCellWidth(state_cell_percent, total_width), "State"),
		generateTableCell(calcCellWidth(name_cell_percent, total_width), "Name"),
		generateTableCell(calcCellWidth(image_cell_percent, total_width), "Image"),
		generateTableCell(calcCellWidth(memory_cell_percent, total_width), "Memory Usage"),
		generateTableCell(calcCellWidth(cpu_cell_percent, total_width), "CPU Usage"),
	)
}

func padResourceUsage(usage string, min_len int) string {
	padding := min_len - len(usage)
	if padding < 0 {
		padding = 0
	}
	return usage + strings.Repeat(" ", padding)
}

func resourceFormatter(use, limit int64, unit string) string {
	return padResourceUsage(fmt.Sprintf("%.2f%s/%.2f%s", float64(use)/float64(1<<30), unit, float64(limit)/float64(1<<30), unit), 17)
}

func generateDataRow(total_width int, datum *docker.ContainerDatum) (elements.StringStyler, error) {
	stats := datum.CachedStats()
	inspect_data := datum.InspectData()
	cpu_usage_percentage := docker.CpuUsagePercentage(&stats.Cpu, &stats.PreCpu, &inspect_data)
	memory_usage_percentage := docker.MemoryUsagePercentage(&stats.Memory)
	memory_usage_str := resourceFormatter(stats.Memory.Usage, stats.Memory.Limit, "GB")
	cpu_usage_str := padResourceUsage(fmt.Sprintf("%.2f%%", docker.CpuUsagePercentage(&stats.Cpu, &stats.PreCpu, &inspect_data)), 8)
	return generateGenericTableRow(
		total_width,
		generateTableCell(calcCellWidth(id_cell_percent, total_width), datum.ID()),
		generateTableCell(calcCellWidth(id_cell_percent, total_width), datum.State()),
		generateTableCell(calcCellWidth(name_cell_percent, total_width), stats.Name),
		generateTableCell(calcCellWidth(image_cell_percent, total_width), datum.Image()),
		elements.PercentageBarDrawer(memory_usage_str,
			memory_usage_percentage,
			calcCellWidth(memory_cell_percent, total_width)-len(memory_usage_str), []rune{},
		),
		elements.PercentageBarDrawer(cpu_usage_str,
			cpu_usage_percentage,
			calcCellWidth(cpu_cell_percent, total_width)-len(cpu_usage_str), []rune{},
		),
	), nil
}

func generateTable(state *tableState) []elements.StringStyler {
	total_width := window.Width(&state.window_state)
	underline_rune := '\u2500'
	offset := 2
	table := make([]elements.StringStyler, len(state.filtered_data)+offset)
	table[0] = generateTableHeader(total_width)
	table[1] = elements.RuneRepeater(underline_rune, tcell.StyleDefault.Foreground(tcell.ColorRebeccaPurple))

	row_ready_ch := make(chan interface{}, len(state.filtered_data))
	defer close(row_ready_ch)
	for index, datum := range state.filtered_data {
		go func(i int, d docker.ContainerDatum) {
			row, err := generateDataRow(total_width, &d)
			if err == nil {
				if d.IsDeleted() {
					table[i+offset] = elements.StrikeThrough(row)
				} else {
					table[i+offset] = row
				}
			} else {
				log.Printf("Got error while generating row: %s\n", err)
			}
			row_ready_ch <- i
		}(index, datum)
	}
	for range state.filtered_data {
		<-row_ready_ch
	}
	return table
}

func calcTableHeight(top, buttom int) int {
	return buttom - top - 4 + 1 - 1
}
