package containers_window

import (
	"dc-top/docker"
	"dc-top/docker/compose"
	"dc-top/gui/view/window"
	"dc-top/gui/view/window/bar_window"
	"fmt"
	"log"

	"github.com/gdamore/tcell/v2"
)

type keyboardMode uint8

const (
	regular keyboardMode = iota
	search
)

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
			err := handleDelete(w.window_context, state)
			if err != nil {
				return err
			}
		}
	case tcell.KeyRune:
		screen := window.GetScreen()
		switch ev.Rune() {
		case 'l':
			if state.focused_id != "" {
				screen.PostEvent(window.NewChangeToLogsWindowEvent(state.focused_id))
			}
		case 'h':
			if state.focused_id != "" {
				screen.PostEvent(window.NewChangeToMainHelpEvent())
			}
		case 'e':
			if state.focused_id != "" {
				index, err := findIndexOfId(state.containers_data.GetData(), state.focused_id)
				if err != nil || state.containers_data.GetData()[index].State() != "running" {
					bar_window.Err([]rune(fmt.Sprintf("Container %s isn't running", state.containers_data.GetData()[index].CachedStats().Name)))
				} else {
					screen.PostEvent(window.NewChangeToContainerShellEvent(state.focused_id))
				}
			}
		case 'v':
			if compose.DcModeEnabled() {
				if !compose.ValidateYaml(w.window_context) {
					bar_window.Err([]rune("docker compose yaml syntax is invalid"))
				} else {
					compose.CreateBackupYaml()
					screen.PostEvent(window.NewChangeToFileEdittorEvent(compose.DcYamlPath(), window.ContainersHolder))
				}
			} else {
				bar_window.Err([]rune("dc mode is disabled"))
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
				handlePause(w.window_context, state.focused_id)
			}
		case 'r':
			if state.focused_id != "" {
				handleRestart(w.window_context, state.focused_id)
			}
		case 's':
			if state.focused_id != "" {
				handleStop(w.window_context, state.focused_id)
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
			bar_window.Info([]rune("Cleared search"))
		case '/':
			state.search_box.Reset()
			bar_window.Info([]rune("Switched to search mode..."))
			state.keyboard_mode = search
		case '!':
			state.is_reverse_sort = !state.is_reverse_sort
		}
	case tcell.KeyF1:
		updateSortType(state, docker.State)
	case tcell.KeyF2:
		updateSortType(state, docker.Name)
	case tcell.KeyF3:
		updateSortType(state, docker.Image)
	case tcell.KeyF4:
		updateSortType(state, docker.Memory)
	case tcell.KeyF5:
		updateSortType(state, docker.Cpu)
	}
	return nil
}

func (state *tableState) searchKeyPress(ev *tcell.EventKey, w *ContainersWindow) {
	key := ev.Key()
	switch key {
	case tcell.KeyEnter:
		state.keyboard_mode = regular
		if state.search_box.Value() != "" {
			bar_window.Info([]rune(fmt.Sprintf("Searching for %s", state.search_box.Value())))
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
	restartIndex(state)
}
