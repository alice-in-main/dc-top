package edittor_window

import (
	"dc-top/docker/compose"
	"dc-top/gui/view/window"
	"dc-top/gui/view/window/bar_window"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
)

type _KeyboardMode uint8

const (
	regular _KeyboardMode = iota
	search
	lookup
)

func (state *edittorState) handleKeypress(ev *tcell.EventKey) {
	switch state.keyboard_mode {
	case regular:
		state.handleRegularKey(ev)
	case search:
		state.handleSearchKey(ev)
	case lookup:
		state.handleLookupKey(ev)
	}
}

func (state *edittorState) handleSearchKey(ev *tcell.EventKey) {
	key := ev.Key()
	switch key {
	case tcell.KeyEscape:
		state.search_box.Reset()
		state.switchToRegular()
	case tcell.KeyCtrlD:
		state.search_box.Reset()
		state.switchToRegular()
	case tcell.KeyEnter:
		state.search_result_row_indices = make([]int, 0)
		for i, row := range state.content {
			if strings.Contains(row, state.search_box.Value()) {
				state.search_result_row_indices = append(state.search_result_row_indices, i)
			}
		}
		if len(state.search_result_row_indices) > 0 {
			state.curr_search_result = 0
			state.focused_line = state.search_result_row_indices[state.curr_search_result]
			state.switchToLookup()
		} else {
			state.switchToRegular()
			bar_window.Err([]rune(fmt.Sprintf("Found 0 results for '%s'", state.search_box.Value())))
		}
	default:
		state.search_box.HandleKey(ev)
	}
}

func (state *edittorState) handleLookupKey(ev *tcell.EventKey) {
	key := ev.Key()
	switch key {
	case tcell.KeyRune:
		r := ev.Rune()
		switch r {
		case 'n':
			if len(state.search_result_row_indices) > 0 {
				num_results := len(state.search_result_row_indices)
				state.curr_search_result = (state.curr_search_result + 1) % num_results
				state.focused_line = state.search_result_row_indices[state.curr_search_result]
				bar_window.Info([]rune(fmt.Sprintf("Showing result %d/%d", state.curr_search_result+1, num_results)))
			}
		case 'N':
			if len(state.search_result_row_indices) > 0 {
				num_results := len(state.search_result_row_indices)
				if state.curr_search_result == 0 {
					state.curr_search_result += num_results
				}
				state.curr_search_result = (state.curr_search_result - 1) % len(state.search_result_row_indices)
				state.focused_line = state.search_result_row_indices[state.curr_search_result]
				bar_window.Info([]rune(fmt.Sprintf("Showing result %d/%d", state.curr_search_result+1, num_results)))
			}
		}
	case tcell.KeyEscape:
		state.switchToRegular()
	case tcell.KeyCtrlD:
		state.switchToRegular()
	case tcell.KeyEnter:
		state.switchToRegular()
	case tcell.KeyCtrlF:
		state.switchToSearch()
	default:
		state.search_box.HandleKey(ev)
	}
}

func (state *edittorState) handleRegularKey(ev *tcell.EventKey) {
	key := ev.Key()
	switch key {
	case tcell.KeyUp:
		state.handleLineFocusChange(state.focused_line - 1)
	case tcell.KeyDown:
		state.handleLineFocusChange(state.focused_line + 1)
	case tcell.KeyLeft:
		state.handleColFocusChange(state.focused_col - 1)
	case tcell.KeyRight:
		state.handleColFocusChange(state.focused_col + 1)
	case tcell.KeyHome:
		state.handleColFocusChange(0)
	case tcell.KeyEnd:
		state.handleColFocusChange(len(state.content[state.focused_line]))
	case tcell.KeyCtrlS:
		state.finalizeEdittor()
	case tcell.KeyCtrlQ:
		window.GetScreen().PostEvent(window.NewReturnUpperViewEvent())
	case tcell.KeyEscape:
		window.GetScreen().PostEvent(window.NewReturnUpperViewEvent())
	case tcell.KeyCtrlH:
		window.GetScreen().PostEvent(window.NewChangeToEdittorHelpEvent())
	case tcell.KeyBackspace2:
		if state.focused_col > 0 {
			state.change_stack.commitTextRemovalChange(string(state.content[state.focused_line][state.focused_col-1]), state.focused_line, state.focused_col-1)
			removeStrFromLine(1, &state.content, state.focused_line, state.focused_col-1)
			state.handleColFocusChange(state.focused_col - 1)
		} else if state.focused_col == 0 {
			if state.focused_line > 0 {
				new_col := len(state.content[state.focused_line-1])
				collapseLine(&state.content, state.focused_line-1)
				state.handleLineFocusChange(state.focused_line - 1)
				state.handleColFocusChange(new_col)
				state.change_stack.commitLineRemove(state.focused_line, state.focused_col)
			}
		}
	case tcell.KeyDelete:
		if state.focused_col < len(state.content[state.focused_line]) {
			state.change_stack.commitTextRemovalChange(string(state.content[state.focused_line][state.focused_col]), state.focused_line, state.focused_col)
			removeStrFromLine(1, &state.content, state.focused_line, state.focused_col)
		} else if state.focused_col == len(state.content[state.focused_line]) {
			if state.focused_line < len(state.content)-1 {
				collapseLine(&state.content, state.focused_line)
				state.change_stack.commitLineRemove(state.focused_line, state.focused_col)
			}
		}
	case tcell.KeyEnter:
		breakLine(&state.content, state.focused_line, state.focused_col)
		state.change_stack.commitLineAdd(state.focused_line, state.focused_col)
		state.handleLineFocusChange(state.focused_line + 1)
		state.handleColFocusChange(0)
	case tcell.KeyRune:
		state.change_stack.commitTextAdditionChange(string(ev.Rune()), state.focused_line, state.focused_col)
		addStrToLine(string(ev.Rune()), &state.content, state.focused_line, state.focused_col)
		state.handleColFocusChange(state.focused_col + 1)
	case tcell.KeyCtrlZ:
		var line, col int
		var err error
		if ev.Modifiers()&tcell.ModAlt == 0 {
			line, col, err = state.change_stack.undoChange(&state.content)
		} else {
			line, col, err = state.change_stack.redoChange(&state.content)
		}
		if err == nil {
			state.handleLineFocusChange(line)
			state.handleColFocusChange(col)
		}
	case tcell.KeyCtrlF:
		state.keyboard_mode = search
	case tcell.KeyCtrlG:
		var line, col int
		if ev.Modifiers()&tcell.ModAlt == 0 {
			line, col = 0, 0
		} else {
			line, col = len(state.content)-1, 0
		}
		state.handleLineFocusChange(line)
		state.handleColFocusChange(col)
	case tcell.KeyCtrlA:
		state.search_box.Reset()
	}
}

func (state *edittorState) finalizeEdittor() {
	if !contentsEquals(state.content, state.original_content) {
		compose.CreateBackupYaml()
		err := writeNewContent(state.content, state.file)
		if err != nil {
			bar_window.Err([]rune(fmt.Sprintf("Got error '%s' while writing to file", err.Error())))
			return
		}
		if !compose.ValidateYaml(state.ctx) {
			output, _ := compose.Config(state.ctx)
			compose.RestoreFromBackup()
			bar_window.Err([]rune("docker-compose yaml contains errors"))
			window.GetScreen().PostEvent(window.NewChangeToErrorEvent(output))
			return
		}

		// Sometimes updating filters fails for unknown reasons so i retry
		for i := 0; i < 3; i++ {
			err = compose.UpdateContainerFilters(state.ctx)
			if err == nil {
				break
			} else {
				log.Printf("Failed to update filters: '%s", err)
				time.Sleep(10 * time.Millisecond)
			}
		}
	}
	window.GetScreen().PostEvent(window.NewReturnUpperViewEvent())
	window.GetScreen().PostEvent(window.NewUpdateDockerCompose())
}

func (state *edittorState) handleLineFocusChange(new_focus_index int) {
	if new_focus_index >= 0 && new_focus_index < len(state.content) {
		state.focused_line = new_focus_index
		if state.focused_col > len(state.content[new_focus_index]) {
			state.handleColFocusChange(len(state.content[new_focus_index]))
		}
	}
}

func (state *edittorState) handleColFocusChange(new_focus_index int) {
	if new_focus_index >= 0 && new_focus_index <= len(state.content[state.focused_line]) {
		state.focused_col = new_focus_index
	} else if new_focus_index < 0 && state.focused_line > 0 {
		state.handleLineFocusChange(state.focused_line - 1)
		state.focused_col = len(state.content[state.focused_line])
	} else if new_focus_index > len(state.content[state.focused_line]) && state.focused_line < len(state.content)-1 {
		state.handleLineFocusChange(state.focused_line + 1)
		state.focused_col = 0
	}
}

func (state *edittorState) switchToRegular() {
	state.keyboard_mode = regular
	bar_window.Info([]rune("Exitted search"))
}

func (state *edittorState) switchToSearch() {
	state.keyboard_mode = search
	bar_window.Info([]rune("Searching..."))
}

func (state *edittorState) switchToLookup() {
	bar_window.Info([]rune("Press n to find next result"))
	state.keyboard_mode = lookup
}
