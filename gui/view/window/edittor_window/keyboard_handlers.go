package edittor_window

import (
	"dc-top/gui/view/window"

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

}

func (state *edittorState) handleLookupKey(ev *tcell.EventKey) {

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
	case tcell.KeyCtrlV:
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
	case tcell.KeyCtrlN:
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
	}
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
