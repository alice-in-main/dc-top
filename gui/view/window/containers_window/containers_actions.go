package containers_window

import (
	docker "dc-top/docker"
	"log"
	"strings"
)

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

func handleRestart(id string) {
	go func(id_to_stop string) {
		if err := docker.RestartContainer(id_to_stop); err != nil {
			log.Printf("Got error '%s' when trying to container delete %s", err, id_to_stop)
			if !strings.Contains(err.Error(), "is already in progress") && !strings.Contains(err.Error(), "No such container") {
				panic(err)
			}
		}
	}(id)
}
