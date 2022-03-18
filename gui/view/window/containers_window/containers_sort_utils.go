package containers_window

import (
	"dc-top/docker"
	"log"
)

func updateSortType(state *tableState, sort_type docker.SortType) {
	if sort_type != docker.None && state.main_sort_type != sort_type {
		state.secondary_sort_type = state.main_sort_type
		state.main_sort_type = sort_type
		log.Print("Updated sort types")
	}
}
