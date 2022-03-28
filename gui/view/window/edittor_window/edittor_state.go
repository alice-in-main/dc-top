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
	is_enabled    bool
	keyboard_mode _KeyboardMode
	content       []string
	top_line      int
	left_col      int
	focused_line  int
	focused_col   int
	change_stack  changeStack
}

func (edittor_window *EdittorWindow) main() error {
	state := edittorState{
		is_enabled:    true,
		keyboard_mode: regular,
		top_line:      0,
		left_col:      0,
		focused_line:  0,
		focused_col:   0,
		change_stack:  newChangeStack(),
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
	edittor_window.drawer_semaphore.Acquire(edittor_window.window_context, 1)
	defer edittor_window.drawer_semaphore.Release(1)
	dimensions := edittor_window.dimensions_generator()

	if state.focused_col > state.left_col+window.Width(&dimensions)-1 {
		state.left_col = state.focused_col - (window.Width(&dimensions) - 1)
	} else if state.focused_col < state.left_col {
		state.left_col = state.focused_col
	}

	if state.focused_line > state.top_line+window.Height(&dimensions)-1 {
		state.top_line = state.focused_line - (window.Height(&dimensions) - 1)
	} else if state.focused_line < state.top_line {
		state.top_line = state.focused_line
	}

	styled_content := make([]elements.StringStyler, window.Height(&dimensions))
	for i := 0; i < window.Height(&dimensions); i++ {
		row_num := i + state.top_line
		if row_num < len(state.content) {
			styled_content[i] = elements.Suffix(elements.HighlightDrawer(state.content[row_num], "", tcell.StyleDefault), state.left_col)
		} else {
			styled_content[i] = elements.RuneNRepeater(rune('~'), 1, tcell.StyleDefault.Foreground(tcell.ColorBlue))
		}
	}

	edittor_drawer := func(x, y int) (rune, tcell.Style) {
		r, s := styled_content[y](x)
		if y+state.top_line == state.focused_line && (x+state.left_col) == state.focused_col {
			s = s.Background(tcell.ColorLightBlue).Foreground(tcell.ColorBlack)
		}
		return r, s
	}

	window.DrawContents(&dimensions, edittor_drawer)
	window.GetScreen().Show()
}
