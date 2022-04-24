package edittor_window

import "dc-top/utils"

type change interface {
	focusedCoords() (line, column int)
}

type textAdditionChange struct {
	text         string
	focused_line int
	focused_col  int
}

func (change *textAdditionChange) focusedCoords() (line, column int) {
	return change.focused_line, change.focused_col
}

func newTextAdditionChange(text string, focused_line int, focused_col int) *textAdditionChange {
	return &textAdditionChange{
		text:         utils.Clone(text),
		focused_line: focused_line,
		focused_col:  focused_col,
	}
}

type textRemovalChange struct {
	text         string
	focused_line int
	focused_col  int
}

func (change *textRemovalChange) focusedCoords() (line, column int) {
	return change.focused_line, change.focused_col
}

func newTextRemovalChange(text string, focused_line int, focused_col int) *textRemovalChange {
	return &textRemovalChange{
		text:         text,
		focused_line: focused_line,
		focused_col:  focused_col,
	}
}

type lineAddChange struct {
	focused_line int
	focused_col  int
}

func (change *lineAddChange) focusedCoords() (line, column int) {
	return change.focused_line, change.focused_col
}

func newLineAdd(focused_line int, focused_col int) *lineAddChange {
	return &lineAddChange{
		focused_line: focused_line,
		focused_col:  focused_col,
	}
}

type lineCollapseChange struct {
	focused_line int
	focused_col  int
}

func (change *lineCollapseChange) focusedCoords() (line, column int) {
	return change.focused_line, change.focused_col
}

func newLineCollapse(focused_line int, focused_col int) *lineCollapseChange {
	return &lineCollapseChange{
		focused_line: focused_line,
		focused_col:  focused_col,
	}
}

type lineDeletionChange struct {
	removed_text string
	focused_line int
	focused_col  int
}

func (change *lineDeletionChange) focusedCoords() (line, column int) {
	return change.focused_line, change.focused_col
}

func newLineDeletion(text string, focused_line int, focused_col int) *lineDeletionChange {
	return &lineDeletionChange{
		removed_text: text,
		focused_line: focused_line,
		focused_col:  focused_col,
	}
}
