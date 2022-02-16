package window

import (
	"dc-top/docker"
	"dc-top/errors"
	"dc-top/gui/elements"
	"dc-top/gui/gui_events"
	"fmt"
	"log"
	"math"
	"strings"

	"github.com/gdamore/tcell/v2"
)

type ContainersWindow struct {
	containers_border_style tcell.Style
	resize_chan             chan interface{}
	next_index_chan         chan interface{}
	prev_index_chan         chan interface{}
	new_index_chan          chan int
	delete_chan             chan interface{}
	switch_to_logs_chan     chan interface{}
	mouse_chan              chan tcell.EventMouse
	new_container_data_chan chan docker.ContainerData
	data_request_chan       chan tableState
	draw_queue              chan tableState
	stop_chan               chan interface{}
}

type tableState struct {
	window_state        WindowState
	index_of_top        int
	table_height        int
	focused_id          string
	containers_data     docker.ContainerData
	main_sort_type      docker.SortType
	secondary_sort_type docker.SortType
}

func NewContainersWindow() ContainersWindow {
	return ContainersWindow{
		containers_border_style: tcell.StyleDefault.Foreground(tcell.Color103),
		resize_chan:             make(chan interface{}),
		next_index_chan:         make(chan interface{}),
		prev_index_chan:         make(chan interface{}),
		new_index_chan:          make(chan int),
		delete_chan:             make(chan interface{}),
		switch_to_logs_chan:     make(chan interface{}),
		mouse_chan:              make(chan tcell.EventMouse),
		new_container_data_chan: make(chan docker.ContainerData),
		data_request_chan:       make(chan tableState),
		stop_chan:               make(chan interface{}),
		draw_queue:              make(chan tableState),
	}
}

func (w *ContainersWindow) Open(s tcell.Screen) {
	s.EnableMouse(tcell.MouseButtonEvents)
	go w.main(s)
}

func (w *ContainersWindow) Resize() {
	w.resize_chan <- nil
}

func (w *ContainersWindow) KeyPress(ev tcell.EventKey) {
	key := ev.Key()
	switch key {
	case tcell.KeyUp:
		w.prev()
	case tcell.KeyDown:
		w.next()
	case tcell.KeyDelete:
		w.delFocused()
	case tcell.KeyRune:
		switch ev.Rune() {
		case 'l':
			w.switchToLogsWindow()
		}
	default:
		return
	}
}

func (w *ContainersWindow) MousePress(ev tcell.EventMouse) {
	w.mouse_chan <- ev
}

func (w *ContainersWindow) Close() {
	w.stop_chan <- nil
}

func (w *ContainersWindow) next() {
	w.next_index_chan <- nil
}

func (w *ContainersWindow) prev() {
	w.prev_index_chan <- nil
}

func (w *ContainersWindow) delFocused() {
	w.delete_chan <- nil
}

func (w *ContainersWindow) switchToLogsWindow() {
	w.switch_to_logs_chan <- nil
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

func (w *ContainersWindow) drawer(stop_drawing chan interface{}) {
	for {
		select {
		case state := <-w.draw_queue:
			log.Printf("Drawing new state...\n")
			DrawBorders(&state.window_state, w.containers_border_style)
			DrawContents(&state.window_state, dockerStatsDrawerGenerator(state))
			state.window_state.Screen.Show()
			log.Printf("Done drawing\n")
		case <-stop_drawing:
			return
		}
	}
}

func (w *ContainersWindow) dockerDataStreamer(stop_streaming chan interface{}) {
	for {
		select {
		case state := <-w.data_request_chan:
			log.Printf("Got request for new data")
			var new_data docker.ContainerData
			if !state.containers_data.AreIdsUpToDate() {
				log.Printf("Ids changed, getting new container stats")
				new_data = docker.GetContainers(&state.containers_data)
			} else {
				state.containers_data.UpdateStats()
				new_data = state.containers_data
			}
			log.Printf("Sending back new data")
			w.new_container_data_chan <- new_data
		case <-stop_streaming:
			return
		}
	}
}

func handleResize(w *ContainersWindow, table_state tableState) tableState {
	log.Printf("Resize request\n")
	x1, y1, x2, y2 := ContainerWindowSize(table_state.window_state.Screen)
	table_state.table_height = y2 - y1 - 4 + 1
	log.Printf("table height is %d\n", table_state.table_height)
	for i, datum := range table_state.containers_data.GetData() {
		if datum.ID() == table_state.focused_id {
			updateIndices(&table_state, i)
			break
		}
	}
	table_state.window_state.SetBorders(x1, y1, x2, y2)
	return table_state
}

func handleNewData(new_data *docker.ContainerData, w *ContainersWindow, table_state tableState) tableState {
	log.Printf("Got new data\n")
	table_state.containers_data = *new_data
	log.Printf("Requesting new data\n")
	if !new_data.Contains(table_state.focused_id) {
		table_state.focused_id = ""
	}
	return table_state
}

func findIndexOfId(data *docker.ContainerData, id string) (int, error) {
	for i, datum := range data.GetData() {
		if datum.ID() == id {
			return i, nil
		}
	}
	return -1, errors.NewNotFoundError("index of id")
}

func handleNewIndex(new_index int, w *ContainersWindow, table_state tableState) tableState {
	if new_index < 0 {
		new_index = table_state.containers_data.Len() - 1
	} else if new_index >= table_state.containers_data.Len() {
		new_index = 0
	}
	table_state.focused_id = table_state.containers_data.GetData()[new_index].ID()
	updateIndices(&table_state, new_index)
	return table_state
}

func handleChangeIndex(is_next bool, w *ContainersWindow, table_state tableState) tableState {
	var new_index int
	log.Printf("Requesting change index\n")
	if table_state.focused_id == "" && table_state.containers_data.Len() > 0 {
		if is_next {
			new_index = 0
		} else {
			new_index = table_state.containers_data.Len() - 1
		}
	} else {
		index, err := findIndexOfId(&table_state.containers_data, table_state.focused_id)
		if err != nil {
			return table_state
		}
		if is_next {
			new_index = index + 1
		} else {
			new_index = index - 1
		}
	}
	return handleNewIndex(new_index, w, table_state)
}

func handleDelete(w *ContainersWindow, table_state tableState) tableState {
	if err := docker.DeleteContainer(table_state.focused_id); err != nil {
		log.Printf("Got error %s when trying to container delete %s", err, table_state.focused_id)
		return table_state
	}
	index, err := findIndexOfId(&table_state.containers_data, table_state.focused_id)
	if err != nil {
		log.Printf("Didn't find id %s in current state after deleting it", table_state.focused_id)
		return table_state
	}
	change_to_next := index != (table_state.containers_data.Len() - 1)
	return handleChangeIndex(change_to_next, w, table_state)
}

func handleMouseEvent(ev *tcell.EventMouse, w *ContainersWindow, table_state tableState) tableState {
	if table_state.window_state.IsOutbounds(ev) {
		x, y := ev.Position()
		log.Printf("outbounds mouse event %d,%d", x, y)
		return table_state
	}
	x, y := table_state.window_state.RelativeMousePosition(ev)
	total_width := table_state.window_state.RightX - table_state.window_state.LeftX
	log.Printf("Handling mouse event that happened on %d, %d", x, y)
	switch {
	case y == 1:
		var new_sort_type docker.SortType = getSortTypeFromMousePress(total_width, x)
		if new_sort_type != docker.None && table_state.main_sort_type != new_sort_type {
			table_state.secondary_sort_type = table_state.main_sort_type
			table_state.main_sort_type = new_sort_type
		}
	case y > 2 && y < table_state.containers_data.Len()+3:
		updateIndices(&table_state, y-3)
		table_state.focused_id = table_state.containers_data.GetData()[y-3].ID()
	}
	return table_state
}

func (w *ContainersWindow) main(s tcell.Screen) {
	x1, y1, x2, y2 := ContainerWindowSize(s)
	window_state := NewWindow(s, x1, y1, x2, y2)
	state := tableState{
		containers_data:     docker.GetContainers(nil),
		index_of_top:        0,
		table_height:        y2 - y1 - 4 + 1,
		window_state:        window_state,
		main_sort_type:      docker.State,
		secondary_sort_type: docker.Name,
	}
	state.containers_data.SortData(state.main_sort_type, state.secondary_sort_type)
	drawer_stop_chan := make(chan interface{})
	go w.drawer(drawer_stop_chan)
	streamer_stop_chan := make(chan interface{})
	go w.dockerDataStreamer(streamer_stop_chan)
	go func() { w.new_container_data_chan <- state.containers_data }()
	for {
		select {
		case <-w.resize_chan:
			state = handleResize(w, state)
		case new_data := <-w.new_container_data_chan:
			state = handleNewData(&new_data, w, state)
			state.containers_data.SortData(state.main_sort_type, state.secondary_sort_type)
			w.data_request_chan <- state
		case <-w.next_index_chan:
			state = handleChangeIndex(true, w, state)
		case <-w.prev_index_chan:
			state = handleChangeIndex(false, w, state)
		case index := <-w.new_index_chan:
			state = handleNewIndex(index, w, state)
		case <-w.delete_chan:
			state = handleDelete(w, state)
		case <-w.switch_to_logs_chan:
			if state.focused_id != "" {
				state.window_state.Screen.PostEvent(gui_events.NewChangeToLogsWindowEvent(state.focused_id))
			}
			continue
		case mouse_event := <-w.mouse_chan:
			state = handleMouseEvent(&mouse_event, w, state)
		case <-w.stop_chan:
			drawer_stop_chan <- nil
			streamer_stop_chan <- nil
			log.Printf("Containers window stopped\n")
			return
		}
		w.draw_queue <- state
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

const (
	id_cell_percent     = 0.04
	state_cell_percent  = 0.04
	name_cell_percent   = 0.12
	image_cell_percent  = 0.24
	memory_cell_percent = 0.28
	cpu_cell_percent    = 0.28
)

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

func getSortTypeFromMousePress(total_width, x int) docker.SortType {
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

func calc_cell_width(relative_size float64, total_width int) int {
	return int(math.Ceil(relative_size * float64(total_width)))
}

func generateTableHeader(total_width int) elements.StringStyler {
	return generateGenericTableRow(
		total_width,
		generateTableCell(calc_cell_width(id_cell_percent, total_width), "ID"),
		generateTableCell(calc_cell_width(state_cell_percent, total_width), "State"),
		generateTableCell(calc_cell_width(name_cell_percent, total_width), "Name"),
		generateTableCell(calc_cell_width(image_cell_percent, total_width), "Image"),
		generateTableCell(calc_cell_width(memory_cell_percent, total_width), "Memory Usage"),
		generateTableCell(calc_cell_width(cpu_cell_percent, total_width), "CPU Usage"),
	)
}

func generateDataRow(total_width int, datum *docker.ContainerDatum) (elements.StringStyler, error) {
	stats := datum.CachedStats()
	cpu_usage_percentage := docker.CpuUsagePercentage(&stats.Cpu, &stats.PreCpu)
	memory_usage_percentage := docker.MemoryUsagePercentage(&stats.Memory)
	resource_formatter := func(use, limit int64, unit string) string {
		return fmt.Sprintf("%.2f%s/%.2f%s ", float64(use)/float64(1<<30), unit, float64(limit)/float64(1<<30), unit)
	}
	return generateGenericTableRow(
		total_width,
		generateTableCell(calc_cell_width(id_cell_percent, total_width), datum.ID()),
		generateTableCell(calc_cell_width(id_cell_percent, total_width), datum.State()),
		generateTableCell(calc_cell_width(name_cell_percent, total_width), stats.Name),
		generateTableCell(calc_cell_width(image_cell_percent, total_width), datum.Image()),
		elements.PercentageBarDrawer(
			resource_formatter(
				stats.Memory.Usage,
				stats.Memory.Limit,
				"GB"),
			memory_usage_percentage,
			calc_cell_width(memory_cell_percent, total_width),
		),
		elements.PercentageBarDrawer(
			fmt.Sprintf("%.2f%% ", cpu_usage_percentage),
			cpu_usage_percentage,
			calc_cell_width(cpu_cell_percent, total_width),
		),
	), nil
}

func generateTable(state *tableState) []elements.StringStyler {
	total_width := (state.window_state.RightX - 1) - (state.window_state.LeftX + 1)
	underline_rune := '\u2500'
	table := make([]elements.StringStyler, state.containers_data.Len()+2)
	table[0] = generateTableHeader(total_width)
	table[1] = elements.RuneRepeater(underline_rune, tcell.StyleDefault.Foreground(tcell.ColorRebeccaPurple))
	offset := 2
	row_ready_ch := make(chan interface{}, state.containers_data.Len())
	defer close(row_ready_ch)
	for index, datum := range state.containers_data.GetData() {
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
				table[i+offset] = elements.RuneRepeater(underline_rune, tcell.StyleDefault)
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
			return rune('\x00'), tcell.StyleDefault
		}
	}
}
