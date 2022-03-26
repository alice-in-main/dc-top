package edittor_window

import "dc-top/utils"

type change interface {
	focusedCoords() (line, column int)
}

type textChange struct {
	original_text     string
	line              int
	prev_focused_line int
	prev_focused_col  int
}

func (change *textChange) focusedCoords() (line, column int) {
	return change.prev_focused_line, change.prev_focused_col
}

func newLineChange(text string, line int, focused_line int, focused_col int) *textChange {
	return &textChange{
		original_text:     utils.Clone(text),
		line:              line,
		prev_focused_line: focused_line,
		prev_focused_col:  focused_col,
	}
}

type lineAddChange struct {
	added_line        int
	added_line_text   string
	prev_focused_line int
	prev_focused_col  int
}

func (change *lineAddChange) focusedCoords() (line, column int) {
	return change.prev_focused_line, change.prev_focused_col
}

func newLineAdd(text string, line_index int, focused_line int, focused_col int) *lineAddChange {
	return &lineAddChange{
		added_line:        line_index,
		added_line_text:   utils.Clone(text),
		prev_focused_line: focused_line,
		prev_focused_col:  focused_col,
	}
}

type lineRemoveChange struct {
	removed_line      int
	removed_line_text string
	prev_focused_line int
	prev_focused_col  int
}

func (change *lineRemoveChange) focusedCoords() (line, column int) {
	return change.prev_focused_line, change.prev_focused_col
}

func newLineRemove(text string, line_index int, focused_line int, focused_col int) *lineRemoveChange {
	return &lineRemoveChange{
		removed_line:      line_index,
		removed_line_text: utils.Clone(text),
		prev_focused_line: focused_line,
		prev_focused_col:  focused_col,
	}
}
