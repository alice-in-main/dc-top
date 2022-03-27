package edittor_window

import (
	"dc-top/gui/elements"
	"dc-top/gui/view/window"
	"io/ioutil"
	"log"
	"strings"

	"github.com/gdamore/tcell/v2"
)

type edittorState struct {
	is_enabled   bool
	content      []string
	top_line     int
	left_col     int
	focused_line int
	focused_col  int
	change_stack changeStack
}

func (edittor_window *EdittorWindow) main() error {
	state := edittorState{
		is_enabled:   true,
		top_line:     0,
		left_col:     0,
		focused_line: 0,
		focused_col:  0,
		change_stack: newChangeStack(),
	}

	raw_bytes, err := ioutil.ReadFile(edittor_window.file.Name())
	if err != nil {
		log.Printf("Failed to read file %s", edittor_window.file.Name())
		return err
	}

	lines := strings.Split(string(raw_bytes), "\n")
	state.content = lines[:len(lines)-1]

	state.draw(edittor_window)

	go func() {
		for {
			select {
			case <-edittor_window.resize_chan:
			case is_enabled := <-edittor_window.enable_toggle:
				state.is_enabled = is_enabled
			case keyboard_ev := <-edittor_window.keyboard_chan:
				state.handleKeypress(keyboard_ev)
			case <-edittor_window.window_context.Done():
				log.Println("edittor is done")
				return
			}
			if state.is_enabled {
				state.draw(edittor_window)
			}
		}
	}()
	return nil
}

func (state *edittorState) draw(edittor_window *EdittorWindow) {
	dimensions := edittor_window.dimensions_generator()

	styled_content := make([]elements.StringStyler, window.Height(&dimensions))
	for i := state.top_line; i < window.Height(&dimensions); i++ {
		if i < len(state.content) {
			styled_content[i] = elements.HighlightDrawer(state.content[i], "", tcell.StyleDefault)
		} else {
			styled_content[i] = elements.RuneNRepeater(rune('~'), 1, tcell.StyleDefault.Foreground(tcell.ColorBlue))
		}
	}

	edittor_drawer := func(x, y int) (rune, tcell.Style) {
		line_index := y + state.top_line
		col_index := x + state.left_col
		r, s := styled_content[line_index](col_index)
		if line_index == state.focused_line && col_index == state.focused_col {
			s = s.Background(tcell.ColorLightBlue).Foreground(tcell.ColorBlack)
		}
		return r, s

	}

	window.DrawContents(&dimensions, edittor_drawer)
	window.GetScreen().Show()
}

func (state *edittorState) handleKeypress(ev *tcell.EventKey) {
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
		state.change_stack.commitLineChange(string(ev.Rune()), state.focused_line, state.focused_col)
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
			state.focused_col = len(state.content[new_focus_index])
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
