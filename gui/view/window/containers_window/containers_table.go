package containers_window

import (
	docker "dc-top/docker"
	"dc-top/gui/elements"
	"fmt"
	"log"
	"math"
	"strings"

	"github.com/gdamore/tcell/v2"
)

const (
	id_cell_percent     = 0.02
	state_cell_percent  = 0.06
	name_cell_percent   = 0.12
	image_cell_percent  = 0.24
	memory_cell_percent = 0.28
	cpu_cell_percent    = 0.28
)

var relative_cell_widths = []float64{
	id_cell_percent,
	state_cell_percent,
	name_cell_percent,
	image_cell_percent,
	memory_cell_percent,
	cpu_cell_percent,
}

func dockerStatsDrawerGenerator(state tableState, window_width int) (func(x, y int) (rune, tcell.Style), error) {
	if state.window_mode == containers {
		data_table := generateTable(&state, window_width)
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
		pretty_info, err := generatePrettyInspectInfo(state, window_width)
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

func generateTableCell(content interface{}) elements.StringStyler {
	switch typed_content := content.(type) {
	case string:
		return elements.TextDrawer(typed_content, tcell.StyleDefault)
	case elements.StringStyler:
		return typed_content
	default:
		log.Println("tried to generate table cell from unknown type")
		panic(1)
	}
}

func generateTableHeader(total_width int, main_sort, secondary_sort docker.SortType, is_reverse_sort bool) []elements.StringStyler {
	const (
		down_arrow = '\u2193'
		up_arrow   = '\u2191'
	)
	var arrow rune = down_arrow
	if is_reverse_sort {
		arrow = up_arrow
	}
	var header_cells = map[docker.SortType]elements.StringStyler{
		docker.State:  elements.TextDrawer(docker.State.String(), tcell.StyleDefault),
		docker.Name:   elements.TextDrawer(docker.Name.String(), tcell.StyleDefault),
		docker.Image:  elements.TextDrawer(docker.Image.String(), tcell.StyleDefault),
		docker.Memory: elements.TextDrawer(docker.Memory.String(), tcell.StyleDefault),
		docker.Cpu:    elements.TextDrawer(docker.Cpu.String(), tcell.StyleDefault),
	}
	header_cells[main_sort] = header_cells[main_sort].Concat(len(main_sort.String()),
		elements.RuneDrawer([]rune{' ', arrow}, tcell.StyleDefault.Foreground(tcell.ColorBlue)))
	header_cells[secondary_sort] = header_cells[secondary_sort].Concat(len(secondary_sort.String()),
		elements.RuneDrawer([]rune{' ', arrow}, tcell.StyleDefault.Foreground(tcell.ColorGray)))

	return []elements.StringStyler{
		generateTableCell("ID"),
		header_cells[docker.State],
		header_cells[docker.Name],
		header_cells[docker.Image],
		header_cells[docker.Memory],
		header_cells[docker.Cpu],
	}
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

func generateDataRow(total_width int, datum *docker.ContainerDatum) []elements.StringStyler {
	stats := datum.CachedStats()
	inspect_data := datum.InspectData()
	cpu_usage_percentage := docker.CpuUsagePercentage(&stats.Cpu, &stats.PreCpu, &inspect_data)
	memory_usage_percentage := docker.MemoryUsagePercentage(&stats.Memory)
	memory_usage_str := resourceFormatter(stats.Memory.Usage, stats.Memory.Limit, "GB")
	cpu_usage_str := padResourceUsage(fmt.Sprintf("%.2f%%", docker.CpuUsagePercentage(&stats.Cpu, &stats.PreCpu, &inspect_data)), 8)
	return []elements.StringStyler{
		generateTableCell(datum.ID()),
		generateTableCell(datum.State()),
		generateTableCell(stats.Name),
		generateTableCell(datum.Image()),
		elements.PercentageBarDrawer(memory_usage_str,
			memory_usage_percentage,
			calcCellWidth(memory_cell_percent, total_width)-len(memory_usage_str), []rune{},
		),
		elements.PercentageBarDrawer(cpu_usage_str,
			cpu_usage_percentage,
			calcCellWidth(cpu_cell_percent, total_width)-len(cpu_usage_str), []rune{},
		),
	}
}

func generateTable(state *tableState, window_width int) []elements.StringStyler {
	var data_rows = make([][]elements.StringStyler, len(state.filtered_data))
	for i, datum := range state.filtered_data {
		data_rows[i] = generateDataRow(window_width, &datum)
	}

	return elements.TableWithHeader(window_width, relative_cell_widths, data_rows, generateTableHeader(window_width, state.main_sort_type, state.secondary_sort_type, state.is_reverse_sort))
}

func calcCellWidth(relative_size float64, total_width int) int {
	return int(math.Ceil(relative_size * float64(total_width)))
}

func calcTableHeight(top, buttom int) int {
	return buttom - top - 4 + 1 - 1
}
