package containers_window

import (
	docker "dc-top/docker"
	"dc-top/gui/elements"
	"log"
)

type tableState struct {
	//common
	is_enabled    bool
	window_mode   windowMode
	keyboard_mode keyboardMode
	focused_id    string
	//containers view
	search_box             elements.TextBox
	index_of_top_container int
	table_height           int
	containers_data        docker.ContainerData
	filtered_data          []docker.ContainerDatum
	main_sort_type         docker.SortType
	secondary_sort_type    docker.SortType
	is_reverse_sort        bool
	top_line_inspect       int
	inspect_height         int
	is_filter_enabled      bool
}

func handleResize(win *ContainersWindow, table_state tableState) tableState {
	log.Printf("Resize request\n")
	restartIndex(&table_state)
	return table_state
}

func handleNewData(new_data *docker.ContainerData, w *ContainersWindow, table_state tableState) tableState {
	// log.Printf("Got new data\n")
	table_state.containers_data = *new_data
	table_state.filtered_data = table_state.containers_data.Filter(table_state.search_box.Value())
	if !new_data.Contains(table_state.focused_id) {
		table_state.focused_id = ""
		table_state.window_mode = containers
	}
	return table_state
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

func restartIndex(state *tableState) {
	index, err := findIndexOfId(state.filtered_data, state.focused_id)
	if err != nil {
		return
	}
	handleNewIndex(index, state)
}

func handleNewIndex(new_index int, table_state *tableState) {
	if len(table_state.filtered_data) == 0 {
		return
	}
	if new_index < 0 {
		new_index = len(table_state.filtered_data) - 1
	} else if new_index >= len(table_state.filtered_data) {
		new_index = 0
	}
	table_state.focused_id = table_state.filtered_data[new_index].ID()
	updateIndices(table_state, new_index)
}

func updateIndices(state *tableState, curr_index int) {
	index_of_buttom := state.index_of_top_container + state.table_height - 1
	if curr_index < state.index_of_top_container {
		state.index_of_top_container = 0
	} else if curr_index >= index_of_buttom {
		state.index_of_top_container = curr_index - state.table_height + 1
	}
	if index_of_buttom > state.containers_data.Len() && state.containers_data.Len() > state.table_height {
		state.index_of_top_container -= (index_of_buttom - state.containers_data.Len() + 1)
	}
	log.Printf("CURR: %d, TOP: %d, BUTTOM: %d\n", curr_index, state.index_of_top_container, index_of_buttom)
}
