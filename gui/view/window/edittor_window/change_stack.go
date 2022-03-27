package edittor_window

import (
	"errors"
	"log"
)

type changeStack struct {
	curr_index int
	changes    []change
}

func newChangeStack() changeStack {
	return changeStack{
		curr_index: 0,
		changes:    make([]change, 0),
	}
}

func (stack *changeStack) commitLineChange(text string, focused_line int, focused_col int) {
	log.Printf("saving %s at %d", text, stack.curr_index)
	stack.changes = append(stack.changes[:stack.curr_index], newTextAdditionChange(text, focused_line, focused_col))
	stack.curr_index++
}

func (stack *changeStack) commitLineAdd(focused_line int, focused_col int) {
	stack.changes = append(stack.changes[:stack.curr_index], newLineAdd(focused_line, focused_col))
	stack.curr_index++
}

func (stack *changeStack) commitLineRemove(focused_line int, focused_col int) {
	stack.changes = append(stack.changes[:stack.curr_index], newLineRemove(focused_line, focused_col))
	stack.curr_index++
}

func (stack *changeStack) undoChange(content *[]string) (line, col int, err error) {
	if stack.curr_index > 0 {
		stack.curr_index--
		_change := stack.changes[stack.curr_index]
		switch _change := _change.(type) {
		case *textAdditionChange:
			removeStrFromLine(len(_change.text), content, _change.focused_line, _change.focused_col)
		case *lineAddChange:
			collapseLine(content, _change.focused_line)
		case *lineRemoveChange:
			breakLine(content, _change.focused_line, _change.focused_col)
		}
		line, col = _change.focusedCoords()
		return line, col, nil
	}
	return -1, -1, errors.New("no more change")
}

func (stack *changeStack) redoChange(content *[]string) (line, col int, err error) {
	if stack.curr_index < len(stack.changes) {
		_change := stack.changes[stack.curr_index]
		switch _change := _change.(type) {
		case *textAdditionChange:
			log.Printf("redoing %s at %d. position %d, %d", _change.text, stack.curr_index, _change.focused_line, _change.focused_col)
			(*content)[_change.focused_line] = (*content)[_change.focused_line][:_change.focused_col] + _change.text + (*content)[_change.focused_line][_change.focused_col:]
		case *lineAddChange:
			breakLine(content, _change.focused_line, _change.focused_col)
		case *lineRemoveChange:
			collapseLine(content, _change.focused_line)
		}
		stack.curr_index++
		line, col = _change.focusedCoords()
		return line, col, nil
	}
	return -1, -1, errors.New("no more change")
}
