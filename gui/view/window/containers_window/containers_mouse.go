package containers_window

import (
	docker "dc-top/docker"
	"dc-top/gui/view/window"
	"log"

	"github.com/gdamore/tcell/v2"
)

func handleMouseEvent(ev *tcell.EventMouse, w *ContainersWindow, table_state tableState) tableState {
	if w.dimensions.IsOutbounds(ev) {
		x, y := ev.Position()
		log.Printf("outbounds mouse event %d,%d", x, y)
		return table_state
	}
	x, y := w.dimensions.RelativeMousePosition(ev)
	total_width := window.Width(&w.dimensions)
	log.Printf("Handling mouse event that happened on %d, %d", x, y)
	switch {
	case y == 1:
		var new_sort_type docker.SortType = getSortTypeFromMousePress(total_width, x)
		if new_sort_type != docker.None && table_state.main_sort_type != new_sort_type {
			table_state.secondary_sort_type = table_state.main_sort_type
			table_state.main_sort_type = new_sort_type
			log.Print("Updated sort types")
		}
	case y > 2 && y < table_state.containers_data.Len()+3:
		i := table_state.index_of_top_container + y - 3
		if i >= table_state.containers_data.Len() {
			break
		}
		updateIndices(&table_state, i)
		table_state.focused_id = table_state.containers_data.GetData()[i].ID()
	}
	return table_state
}

func getSortTypeFromMousePress(total_width, x int) docker.SortType {
	var (
		cell_to_sort_type = map[int]docker.SortType{
			0: docker.None,
			1: docker.State,
			2: docker.Name,
			3: docker.Image,
			4: docker.Memory,
			5: docker.Cpu,
		}
	)
	widths := getCellWidths(total_width)
	for i, cummulative_size := 0, 0; i < len(widths); i++ {
		next_cummulative_size := cummulative_size + widths[i]
		if x >= cummulative_size && x < next_cummulative_size {
			return cell_to_sort_type[i]
		}
		cummulative_size = next_cummulative_size
	}
	return docker.None
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
