package window

// TODO: add start date in inspect mode
// TODO: possible for slow sort is copying slices
// TODO: add arrows for sorted by

import (
	"context"
	docker "dc-top/docker"
	"dc-top/docker/compose"
	"dc-top/errors"
	"dc-top/gui/elements"
	"fmt"
	"log"
	"math"
	"sort"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/go-connections/nat"
	"github.com/gdamore/tcell/v2"
)

type windowMode uint8

const (
	containers windowMode = iota
	inspect
)

type keyboardMode uint8

const (
	regular keyboardMode = iota
	search
)

type ContainersWindow struct {
	//common
	cached_state            tableState
	resize_chan             chan interface{}
	new_container_data_chan chan docker.ContainerData
	data_request_chan       chan tableState
	draw_queue              chan tableState
	enable_toggle           chan bool
	stop_chan               chan interface{}
	//containers view
	mouse_chan    chan tcell.EventMouse
	keyboard_chan chan tcell.EventKey
}

type tableState struct {
	//common
	window_state WindowState
	focused_id   string
	//containers view
	is_enabled    bool
	window_mode   windowMode
	keyboard_mode keyboardMode
	search_box    elements.TextBox
	// search_buffer          string
	// search_buffer_index    int
	index_of_top_container int
	table_height           int
	containers_data        docker.ContainerData
	filtered_data          []docker.ContainerDatum
	main_sort_type         docker.SortType
	secondary_sort_type    docker.SortType
	top_line_inspect       int
	inspect_height         int
}

func NewContainersWindow() ContainersWindow {
	return ContainersWindow{
		//common
		resize_chan:             make(chan interface{}),
		draw_queue:              make(chan tableState),
		enable_toggle:           make(chan bool),
		stop_chan:               make(chan interface{}),
		new_container_data_chan: make(chan docker.ContainerData),
		//containers view
		mouse_chan:        make(chan tcell.EventMouse),
		keyboard_chan:     make(chan tcell.EventKey),
		data_request_chan: make(chan tableState),
	}
}

func (w *ContainersWindow) Open() {
	go w.main()
}

func (w *ContainersWindow) Resize() {
	w.resize_chan <- nil
}

func (w *ContainersWindow) KeyPress(ev tcell.EventKey) {
	w.keyboard_chan <- ev
}

func (w *ContainersWindow) MousePress(ev tcell.EventMouse) {
	w.mouse_chan <- ev
}

type getTotalStats struct{}
type totalStatsSummary struct {
	totalCpuUsage       int64
	totalSystemCpuUsage int64
	totalMemUsage       int64
}

func (w *ContainersWindow) HandleEvent(ev interface{}, sender WindowType) (interface{}, error) {
	switch ev := ev.(type) {
	case getTotalStats:
		var total_cpu_usage int64
		var total_mem_usage int64
		for _, datum := range w.cached_state.containers_data.GetData() {
			total_cpu_usage += datum.CachedStats().Cpu.ContainerUsage.TotalUsage - datum.CachedStats().PreCpu.ContainerUsage.TotalUsage
			total_mem_usage += datum.CachedStats().Memory.Usage
		}
		var system_cpu_usage int64
		if w.cached_state.containers_data.Len() == 0 {
			system_cpu_usage = 99999999999999999
		} else {
			system_cpu_usage = w.cached_state.containers_data.GetData()[0].CachedStats().Cpu.SystemUsage - w.cached_state.containers_data.GetData()[0].CachedStats().PreCpu.SystemUsage
		}
		summary := totalStatsSummary{
			totalCpuUsage:       total_cpu_usage,
			totalSystemCpuUsage: system_cpu_usage,
			totalMemUsage:       total_mem_usage,
		}
		GetScreen().PostEvent(NewMessageEvent(sender, ContainersHolder, summary))
	default:
		log.Fatal("Got unknown event in holder", ev)
	}
	return nil, nil
}

func (w *ContainersWindow) Disable() {
	log.Printf("Disable containers...")
	w.enable_toggle <- false
}

func (w *ContainersWindow) Enable() {
	log.Printf("Enable containers...")
	w.enable_toggle <- true
}

func (w *ContainersWindow) Close() {
	w.stop_chan <- nil
}

func (w *ContainersWindow) drawer(c context.Context) {
	for {
		select {
		case state := <-w.draw_queue:
			if state.is_enabled {
				DrawBorders(&state.window_state)
				drawer_func, err := dockerStatsDrawerGenerator(state)
				if err != nil {
					log.Printf("Got error %s while drawing\n", err)
				}
				DrawContents(&state.window_state, drawer_func)
				GetScreen().Show()
			}
		case <-c.Done():
			log.Printf("Containers window stopped drwaing...\n")
			return
		}
	}
}

func (w *ContainersWindow) dockerDataStreamer(c context.Context) {
	for {
		select {
		case state := <-w.data_request_chan:
			var new_data docker.ContainerData
			up_to_date, err := state.containers_data.AreIdsUpToDate()
			exitIfErr(err)
			if !up_to_date {
				log.Printf("Ids changed, getting new container stats")
				new_data, err = docker.GetContainers(&state.containers_data)
				exitIfErr(err)
			} else {
				state.containers_data.UpdateStats()
				new_data = state.containers_data
			}
			select {
			case <-c.Done():
				log.Printf("Stopped streaming containers data 1")
				return
			default:
				log.Printf("Sending back new data")
				w.new_container_data_chan <- new_data
			}
		case <-c.Done():
			log.Printf("Stopped streaming containers data 2")
			return
		}
	}
}

func handleResize(w *ContainersWindow, table_state tableState) tableState {
	log.Printf("Resize request\n")
	x1, y1, x2, y2 := ContainerWindowSize()
	table_state.table_height = calcTableHeight(y1, y2)
	table_state.inspect_height = y2 - y1 - 2 + 1
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
	table_state.filtered_data = table_state.containers_data.Filter(table_state.search_box.Value())
	if !new_data.Contains(table_state.focused_id) {
		table_state.focused_id = ""
		table_state.window_mode = containers
	}
	return table_state
}

func handleNewIndex(new_index int, table_state *tableState) {
	if new_index < 0 {
		new_index = len(table_state.filtered_data) - 1
	} else if new_index >= len(table_state.filtered_data) {
		new_index = 0
	}
	table_state.focused_id = table_state.filtered_data[new_index].ID()
	updateIndices(table_state, new_index)
}

func handleChangeIndex(is_next bool, table_state *tableState) {
	var new_index int
	log.Printf("Requesting change index\n")
	if table_state.focused_id == "" && len(table_state.filtered_data) > 0 {
		if is_next {
			new_index = 0
		} else {
			new_index = len(table_state.filtered_data) - 1
		}
	} else {
		index, err := findIndexOfId(table_state.filtered_data, table_state.focused_id)
		if err != nil {
			if table_state.containers_data.Len() == 0 {
				return
			}
			index = 0
		}
		if is_next {
			new_index = index + 1
		} else {
			new_index = index - 1
		}
	}
	handleNewIndex(new_index, table_state)
}

func handleDelete(table_state *tableState) error {
	index, err := findIndexOfId(table_state.containers_data.GetData(), table_state.focused_id)
	if err != nil {
		return err
	}
	go func(id_to_delete string) {
		if err := docker.DeleteContainer(id_to_delete); err != nil {
			log.Printf("Got error '%s' when trying to container delete %s", err, id_to_delete)
			if !strings.Contains(err.Error(), "is already in progress") && !strings.Contains(err.Error(), "No such container") {
				panic(err)
			}
		}
	}(table_state.focused_id)
	change_to_next := index != (table_state.containers_data.Len() - 1)
	handleChangeIndex(change_to_next, table_state)
	return nil
}

func handlePause(id string) {
	go func(id_to_pause string) {
		if err := docker.PauseContainer(id_to_pause); err != nil {
			log.Printf("Got error '%s' when trying to container delete %s", err, id_to_pause)
			if strings.Contains(err.Error(), "is already paused") {
				if err := docker.UnpauseContainer(id_to_pause); err != nil {
					panic(err)
				}
			} else if !strings.Contains(err.Error(), "is already in progress") &&
				!strings.Contains(err.Error(), "No such container") &&
				!strings.Contains(err.Error(), "is not running") {
				panic(err)
			}
		}
	}(id)
}

func handleStop(id string) {
	go func(id_to_stop string) {
		if err := docker.StopContainer(id_to_stop); err != nil {
			log.Printf("Got error '%s' when trying to container delete %s", err, id_to_stop)
			if !strings.Contains(err.Error(), "is already in progress") && !strings.Contains(err.Error(), "No such container") {
				panic(err)
			}
		}
	}(id)
}

func handleMouseEvent(ev *tcell.EventMouse, w *ContainersWindow, table_state tableState) tableState {
	if table_state.window_state.IsOutbounds(ev) {
		x, y := ev.Position()
		log.Printf("outbounds mouse event %d,%d", x, y)
		return table_state
	}
	x, y := table_state.window_state.RelativeMousePosition(ev)
	total_width := Width(&table_state.window_state)
	log.Printf("Handling mouse event that happened on %d, %d", x, y)
	switch {
	case y == 1:
		var new_sort_type docker.SortType = getSortTypeFromMousePress(total_width, x)
		if new_sort_type != docker.None && table_state.main_sort_type != new_sort_type {
			table_state.secondary_sort_type = table_state.main_sort_type
			table_state.main_sort_type = new_sort_type
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

func handleKeyboardEvent(ev *tcell.EventKey, w *ContainersWindow, table_state tableState) (tableState, error) {
	if table_state.keyboard_mode == regular {
		err := table_state.regularKeyPress(ev, w)
		if err != nil {
			return table_state, err
		}
	} else if table_state.keyboard_mode == search {
		table_state.searchKeyPress(ev, w)
	} else {
		log.Fatal("Unknown keyboard mode", table_state.keyboard_mode)
	}
	return table_state, nil
}

func (state *tableState) regularKeyPress(ev *tcell.EventKey, w *ContainersWindow) error {
	key := ev.Key()
	switch key {
	case tcell.KeyUp:
		if state.window_mode == containers {
			handleChangeIndex(false, state)
		} else if state.window_mode == inspect {
			state.top_line_inspect--
		}
	case tcell.KeyDown:
		if state.window_mode == containers {
			handleChangeIndex(true, state)
		} else if state.window_mode == inspect {
			state.top_line_inspect++
		}
	case tcell.KeyDelete:
		state.window_mode = containers
		if state.focused_id != "" {
			err := handleDelete(state)
			if err != nil {
				return err
			}
		}
	case tcell.KeyRune:
		screen := GetScreen()
		switch ev.Rune() {
		case 'l':
			if state.focused_id != "" {
				screen.PostEvent(NewChangeToLogsWindowEvent(state.focused_id))
			}
		case 'e':
			if state.focused_id != "" {
				screen.PostEvent(NewChangeToContainerShellEvent(state.focused_id))
			}
		case 'v':
			if compose.DcModeEnabled() {
				if !compose.ValidateYaml(context.TODO()) {
					screen.PostEvent(NewMessageEvent(Bar, ContainersHolder, ErrorMessage{Msg: []rune("docker compose yaml syntax is invalid")}))
				} else {
					compose.CreateBackupYaml()
					screen.PostEvent(NewChangeToFileEdittorEvent(compose.DcYamlPath(), ContainersHolder))
				}
			} else {
				screen.PostEvent(NewMessageEvent(Bar, ContainersHolder, ErrorMessage{Msg: []rune("dc mode is disabled")}))
			}
		case 'i':
			if state.window_mode == containers {
				_, err := findIndexOfId(state.containers_data.GetData(), state.focused_id)
				if err != nil {
					return err
				}
				state.window_mode = inspect
			} else {
				state.window_mode = containers
			}
			log.Println("Toggling inspect mode")
		case 'p':
			if state.focused_id != "" {
				handlePause(state.focused_id)
			}
		case 's':
			if state.focused_id != "" {
				handleStop(state.focused_id)
			}
		case 'g':
			if state.window_mode == containers {
				handleNewIndex(0, state)
			} else if state.window_mode == inspect {
				state.top_line_inspect = 0
			}
		case 'G':
			if state.window_mode == containers {
				handleNewIndex(state.containers_data.Len()-1, state)
			}
		case 'c':
			state.search_box.Reset()
			state.filtered_data = state.containers_data.Filter(state.search_box.Value())
			screen.PostEvent(NewMessageEvent(Bar, ContainersHolder, InfoMessage{Msg: []rune("Cleared search")}))
		case '/':
			state.search_box.Reset()
			screen.PostEvent(NewMessageEvent(Bar, ContainersHolder, InfoMessage{Msg: []rune("Switched to search mode...")}))
			state.keyboard_mode = search
		}
	}
	return nil
}

func (state *tableState) searchKeyPress(ev *tcell.EventKey, w *ContainersWindow) {
	key := ev.Key()
	switch key {
	case tcell.KeyEnter:
		state.keyboard_mode = regular
		if state.search_box.Value() != "" {
			GetScreen().PostEvent(NewMessageEvent(Bar, ContainersHolder, InfoMessage{Msg: []rune(fmt.Sprintf("Searching for %s", state.search_box.Value()))}))
		}
	case tcell.KeyEscape:
		state.keyboard_mode = regular
		state.search_box.Reset()
	case tcell.KeyCtrlD:
		state.keyboard_mode = regular
		state.search_box.Reset()
	default:
		state.search_box.HandleKey(ev)
	}
	state.filtered_data = state.containers_data.Filter(state.search_box.Value())
}

func calcTableHeight(top, buttom int) int {
	return buttom - top - 4 + 1 - 1
}

func (w *ContainersWindow) main() {
	x1, y1, x2, y2 := ContainerWindowSize()
	window_state := NewWindow(x1, y1, x2, y2)
	data, err := docker.GetContainers(nil)
	exitIfErr(err)
	state := tableState{
		is_enabled:      true,
		containers_data: data,
		filtered_data:   data.Filter(""),
		search_box: elements.NewTextBox(
			elements.TextDrawer(" /", tcell.StyleDefault.Foreground(tcell.ColorYellow)),
			2,
			tcell.StyleDefault,
			tcell.StyleDefault.Background(tcell.ColorWhite).Foreground(tcell.ColorBlack)),
		index_of_top_container: 0,
		table_height:           calcTableHeight(y1, y2),
		window_state:           window_state,
		main_sort_type:         docker.State,
		secondary_sort_type:    docker.Name,
		window_mode:            containers,
		keyboard_mode:          regular,
		top_line_inspect:       0,
		inspect_height:         y2 - y1 - 2 + 1,
	}
	state.containers_data.SortData(state.main_sort_type, state.secondary_sort_type)
	window_context, cancel := context.WithCancel(context.TODO())
	w.cached_state = state
	go w.drawer(window_context)
	go w.dockerDataStreamer(window_context)
	go func() { w.new_container_data_chan <- state.containers_data }()
	for {
		select {
		case state.is_enabled = <-w.enable_toggle:
			log.Printf("changed enabled to %t in containers holder", state.is_enabled)
		case <-w.resize_chan:
			state = handleResize(w, state)
		case new_data := <-w.new_container_data_chan:
			state = handleNewData(&new_data, w, state)
			w.cached_state = state
			state.containers_data.SortData(state.main_sort_type, state.secondary_sort_type)
			w.data_request_chan <- state
		case mouse_event := <-w.mouse_chan:
			log.Println("Handling mouse event")
			state = handleMouseEvent(&mouse_event, w, state)
		case keyboard_event := <-w.keyboard_chan:
			state, err = handleKeyboardEvent(&keyboard_event, w, state)
			exitIfErr(err)
		case <-w.stop_chan:
			log.Printf("Stopping all containers window routines\n")
			cancel()
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
	total_width := Width(&state.window_state)
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

func generateResourceUsageStyler(usage, quota, limit int64, resource, unit string, bar_len int) elements.StringStyler {
	var quota_desc string
	if quota == 0 {
		quota_desc = " Quota isn't set"
		quota = limit
	} else {
		quota_desc = fmt.Sprintf(" Quota: %.2f%s", float64(quota)/float64(1<<30), unit)
	}
	usage_human_readable := resourceFormatter(usage, quota, unit)
	return elements.ValuesBarDrawer(
		resource,
		0.0,
		float64(quota),
		float64(usage),
		bar_len,
		[]rune(" "+usage_human_readable+quota_desc))
}

func generatePortMap(ports nat.PortMap) []string {
	var port_map []string = make([]string, len(ports))
	index := 0
	for port, port_bindings := range ports {
		if len(port_bindings) > 0 {
			for _, binding := range port_bindings {
				port_map[index] = fmt.Sprintf("  %s : %s", port, binding.HostPort)
			}
		} else {
			port_map[index] = fmt.Sprintf("  %s not mapped", port)
		}
		index++
	}
	sort.Strings(sort.StringSlice(port_map))
	return port_map
}

func generateMountsMap(mounts []types.MountPoint) []string {
	sort.SliceStable(mounts, func(i, j int) bool { return mounts[i].Destination < mounts[j].Destination })
	var parsed_mounts []string = make([]string, 3*len(mounts))
	for i := 0; i < 2*len(mounts); i += 3 {
		mount_num := i / 3
		parsed_mounts[i] = fmt.Sprintf("  %s> %s", mounts[mount_num].Type, mounts[mount_num].Name)
		parsed_mounts[i+1] = fmt.Sprintf("    %s:%s", mounts[mount_num].Source, mounts[mount_num].Destination)
		parsed_mounts[i+2] = fmt.Sprintf("    Mode: %s, Driver: %s, RW: %t", mounts[mount_num].Mode, mounts[mount_num].Driver, mounts[mount_num].RW)
	}
	return parsed_mounts
}

func generateNetworkUsage(curr_stats map[string]docker.NetworkUsage, prev_stats map[string]docker.NetworkUsage) ([]elements.StringStyler, error) {
	ret := make([]elements.StringStyler, 0)
	for network_interface, usage := range curr_stats {
		ret = append(ret, elements.TextDrawer("  "+network_interface, tcell.StyleDefault))
		mapped_usage, err := docker.NetworkUsageToMapOfInt(usage)
		if err != nil {
			return nil, err
		}
		prev_mapped_usage, err := docker.NetworkUsageToMapOfInt(prev_stats[network_interface])
		if err != nil {
			return nil, err
		}
		time_diff := curr_stats[network_interface].LastUpdateTime.Sub(prev_stats[network_interface].LastUpdateTime).Seconds()

		sorted_usage_keys := make([]string, 0, len(mapped_usage))
		for k := range mapped_usage {
			sorted_usage_keys = append(sorted_usage_keys, k)
		}
		sort.Strings(sorted_usage_keys)

		for _, key := range sorted_usage_keys {
			max_line_len := 30
			curr_usage := mapped_usage[key]
			prev_usage := prev_mapped_usage[key]
			line := fmt.Sprintf("    %s:%d", key, curr_usage)
			padding_len := max_line_len - len(line)
			line += strings.Repeat(" ", padding_len)
			if strings.Contains(key, "byte") {
				speed := float64(curr_usage-prev_usage) / time_diff
				var unit string
				switch {
				case speed > (1 << 30):
					speed /= (1 << 30)
					unit = "GB/s"
				case speed > (1 << 20):
					speed /= (1 << 20)
					unit = "MB/s"
				case speed > (1 << 10):
					speed /= (1 << 10)
					unit = "KB/s"
				default:
					unit = "Bytes/s"
				}
				line += fmt.Sprintf("%.3f%s", speed, unit)
			} else {
				line += fmt.Sprintf("%.3f/s", float64(curr_usage-prev_usage)/time_diff)
			}
			ret = append(ret, elements.TextDrawer(line, tcell.StyleDefault))
		}

	}
	return ret, nil
}

func generateInspectSeperator() elements.StringStyler {
	const underline_rune = '\u2500'
	return elements.RuneRepeater(underline_rune, tcell.StyleDefault.Foreground(tcell.ColorPaleGreen))
}

func generatePrettyInspectInfo(state tableState) (map[int]elements.StringStyler, error) {
	var info map[int]elements.StringStyler = make(map[int]elements.StringStyler)
	var info_arr []elements.StringStyler = make([]elements.StringStyler, 0)

	index, err := findIndexOfId(state.containers_data.GetData(), state.focused_id)
	if err != nil {
		return info, fmt.Errorf("didn't find container '%s' for inspecting", state.focused_id)
	}
	stats := state.containers_data.GetData()[index]
	inspect_info := stats.InspectData()
	window_width := Width(&state.window_state)

	if inspect_info.ContainerJSONBase == nil {
		return info, fmt.Errorf("inspection ended")
	}

	info_arr = append(info_arr,
		elements.TextDrawer("Name: "+stats.CachedStats().Name, tcell.StyleDefault),
		elements.TextDrawer("ID: "+stats.ID(), tcell.StyleDefault),
		elements.TextDrawer("Image: "+stats.Image(), tcell.StyleDefault),
		elements.TextDrawer("State: "+stats.State(), tcell.StyleDefault),
		elements.TextDrawer(fmt.Sprintf("Restart count: %d", inspect_info.RestartCount), tcell.StyleDefault),
	)
	cpu_usage := stats.CachedStats().Cpu.ContainerUsage.TotalUsage - stats.CachedStats().PreCpu.ContainerUsage.TotalUsage
	cpu_quota := inspect_info.HostConfig.NanoCPUs
	cpu_limit := stats.CachedStats().Cpu.SystemUsage - stats.CachedStats().PreCpu.SystemUsage
	memory_usage := stats.CachedStats().Memory.Usage
	memory_quota := inspect_info.HostConfig.Memory
	memory_limit := stats.CachedStats().Memory.Limit
	max_desc_len := 25
	bar_len := window_width - max_desc_len
	if bar_len < 0 {
		bar_len = 0
	} else if bar_len > 40 {
		bar_len = 40
	}
	info_arr = append(info_arr,
		generateResourceUsageStyler(cpu_usage, cpu_quota, cpu_limit, "CPU:    ", "Cores", bar_len),
		generateResourceUsageStyler(memory_usage, memory_quota, memory_limit, "Memory: ", "GB", bar_len),
		generateInspectSeperator(),
		elements.TextDrawer("Ports:", tcell.StyleDefault),
	)
	port_map := generatePortMap(inspect_info.NetworkSettings.Ports)
	for _, port_binding := range port_map {
		info_arr = append(info_arr, elements.TextDrawer(port_binding, tcell.StyleDefault))
	}
	info_arr = append(info_arr, generateInspectSeperator(),
		elements.TextDrawer("Mounts:", tcell.StyleDefault),
	)
	parsed_mounts := generateMountsMap(inspect_info.Mounts)
	for _, mount := range parsed_mounts {
		info_arr = append(info_arr, elements.TextDrawer(mount, tcell.StyleDefault))
	}
	info_arr = append(info_arr, generateInspectSeperator(),
		elements.TextDrawer("Network Usage:", tcell.StyleDefault),
	)
	network_usage, err := generateNetworkUsage(stats.CachedStats().Network, stats.CachedStats().PreNetwork)
	if err != nil {
		return info, err
	}
	info_arr = append(info_arr, network_usage...)
	var row_offset int
	num_rows := len(info_arr)
	if num_rows > state.inspect_height {
		row_offset += state.top_line_inspect % (1 + num_rows - state.inspect_height)
		if row_offset < 0 {
			row_offset += 1 + num_rows - state.inspect_height
		}
	}
	for i, line := range info_arr {
		info[i-row_offset] = line
	}
	return info, nil
}

func updateIndices(state *tableState, curr_index int) {
	index_of_buttom := state.index_of_top_container + state.table_height - 1
	if curr_index < state.index_of_top_container {
		state.index_of_top_container = curr_index
	} else if curr_index >= index_of_buttom {
		state.index_of_top_container = curr_index - state.table_height + 1
	}
	if index_of_buttom > state.containers_data.Len() && state.containers_data.Len() > state.table_height {
		state.index_of_top_container -= (index_of_buttom - state.containers_data.Len() + 1)
	}
	log.Printf("CURR: %d, TOP: %d, BUTTOM: %d\n", curr_index, state.index_of_top_container, index_of_buttom)
}

func findIndexOfId(data []docker.ContainerDatum, id string) (int, error) {
	for i, datum := range data {
		if datum.ID() == id {
			return i, nil
		}
	}
	return -1, errors.NewNotFoundError("index of id")
}
