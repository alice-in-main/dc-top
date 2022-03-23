package containers_window

import (
	"context"
	docker "dc-top/docker"
	"dc-top/gui/elements"
	"dc-top/gui/view/window"
	"log"
	"time"

	"github.com/gdamore/tcell/v2"
	"golang.org/x/sync/semaphore"
)

// TODO: random log file name + log levels
// TODO: finish dc
// TODO: fix containers scrolling (after deletion and last container with search)

type windowMode uint8

const (
	containers windowMode = iota
	inspect
)

type ContainersWindow struct {
	window_context   context.Context
	window_cancel    context.CancelFunc
	drawer_semaphore *semaphore.Weighted
	//common
	dimensions_generator    func() window.Dimensions
	cached_state            tableState
	resize_chan             chan interface{}
	new_container_data_chan chan docker.ContainerData
	data_request_chan       chan tableState
	draw_queue              chan tableState
	enable_toggle           chan bool
	//containers view
	mouse_chan    chan tcell.EventMouse
	keyboard_chan chan tcell.EventKey
}

func NewContainersWindow() ContainersWindow {
	return ContainersWindow{
		//common
		dimensions_generator: func() window.Dimensions {
			x1, y1, x2, y2 := window.ContainerWindowSize()
			return window.NewDimensions(x1, y1, x2, y2, true)
		},
		drawer_semaphore:        semaphore.NewWeighted(1),
		resize_chan:             make(chan interface{}),
		draw_queue:              make(chan tableState),
		enable_toggle:           make(chan bool),
		new_container_data_chan: make(chan docker.ContainerData),
		//containers view
		mouse_chan:        make(chan tcell.EventMouse),
		keyboard_chan:     make(chan tcell.EventKey),
		data_request_chan: make(chan tableState),
	}
}

func (w *ContainersWindow) Open(view_ctx context.Context) {
	log.Println("Opening containers")
	w.window_context, w.window_cancel = context.WithCancel(view_ctx)
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

type GetTotalStats struct{}
type TotalStatsSummary struct {
	TotalCpuUsage       int64
	TotalSystemCpuUsage int64
	TotalMemUsage       int64
}

func (w *ContainersWindow) HandleEvent(ev interface{}, sender window.WindowType) (interface{}, error) {
	switch ev := ev.(type) {
	case GetTotalStats:
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
		summary := TotalStatsSummary{
			TotalCpuUsage:       total_cpu_usage,
			TotalSystemCpuUsage: system_cpu_usage,
			TotalMemUsage:       total_mem_usage,
		}
		window.GetScreen().PostEvent(window.NewMessageEvent(sender, window.ContainersHolder, summary))
	default:
		log.Fatal("Got unknown event in holder", ev)
	}
	return nil, nil
}

func (w *ContainersWindow) Disable() {
	log.Printf("Disable containers...")
	w.enable_toggle <- false
	w.drawer_semaphore.Acquire(w.window_context, 1)
	w.drawer_semaphore.Release(1)
}

func (w *ContainersWindow) Enable() {
	log.Printf("Enable containers...")
	w.enable_toggle <- true
}

func (w *ContainersWindow) Close() {
	w.window_cancel()
}

func (w *ContainersWindow) main() {
	_, y1, _, y2 := window.ContainerWindowSize()
	data, err := docker.NewContainerData(w.window_context)
	window.ExitIfErr(err)
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
		main_sort_type:         docker.State,
		secondary_sort_type:    docker.Name,
		is_reverse_sort:        false,
		window_mode:            containers,
		keyboard_mode:          regular,
		top_line_inspect:       0,
		inspect_height:         y2 - y1 - 2 + 1,
	}
	state.containers_data = data.GetSortedData(state.main_sort_type, state.secondary_sort_type, false)
	w.cached_state = state
	go w.drawer()
	go w.dockerDataStreamer()
	go w.sendInitialStateAsync(&state)
	for {
		select {
		case state.is_enabled = <-w.enable_toggle:
			log.Printf("changed enabled to %t in containers holder", state.is_enabled)
		case <-w.resize_chan:
			state = handleResize(w, state)
		case new_data := <-w.new_container_data_chan:
			new_data = new_data.GetSortedData(state.main_sort_type, state.secondary_sort_type, state.is_reverse_sort)
			state = handleNewData(&new_data, w, state)
			w.cached_state = state
			w.data_request_chan <- state
		case mouse_event := <-w.mouse_chan:
			log.Println("Handling mouse event")
			state = handleMouseEvent(&mouse_event, w, state)
			state.containers_data = state.containers_data.GetSortedData(state.main_sort_type, state.secondary_sort_type, state.is_reverse_sort)
			state.filtered_data = state.containers_data.Filter(state.search_box.Value())
		case keyboard_event := <-w.keyboard_chan:
			state, err = handleKeyboardEvent(&keyboard_event, w, state)
			window.ExitIfErr(err)
			state.containers_data = state.containers_data.GetSortedData(state.main_sort_type, state.secondary_sort_type, state.is_reverse_sort)
			state.filtered_data = state.containers_data.Filter(state.search_box.Value())
		case <-w.window_context.Done():
			log.Printf("Stopping all containers window routines\n")
			return
		}
		select {
		case w.draw_queue <- state:
		case <-w.window_context.Done():
			return
		}
	}
}

func (w *ContainersWindow) drawer() {
	for {
		select {
		case state := <-w.draw_queue:
			if state.is_enabled {
				w.drawer_semaphore.Acquire(w.window_context, 1)
				dimensions := w.dimensions_generator()
				drawer_func, err := dockerStatsDrawerGenerator(state, window.Width(&dimensions))
				if err != nil {
					log.Printf("Got error %s while drawing\n", err)
				}
				window.DrawContents(&dimensions, drawer_func)
				window.GetScreen().Show()
				w.drawer_semaphore.Release(1)
			}
		case <-w.window_context.Done():
			log.Printf("Containers window stopped drwaing...\n")
			return
		}
	}
}

func (w *ContainersWindow) dockerDataStreamer() {
	for {
		select {
		case state := <-w.data_request_chan:
			var new_data docker.ContainerData
			new_data, err := docker.UpdatedContainerData(w.window_context, &state.containers_data)
			window.ExitIfErr(err)
			if new_data.Len() == 0 {
				time.Sleep(200 * time.Millisecond)
			}
			select {
			case <-w.window_context.Done():
				log.Printf("Stopped streaming containers data 1")
				return
			default:
				log.Printf("Sending back new data")
				w.new_container_data_chan <- new_data
			}
		case <-w.window_context.Done():
			log.Printf("Stopped streaming containers data 2")
			return
		}
	}
}

func (w *ContainersWindow) sendInitialStateAsync(state *tableState) {
	select {
	case <-w.window_context.Done():
		return
	case w.new_container_data_chan <- state.containers_data:
	}
}
