package edittor_window

import "log"

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

func (stack *changeStack) commitLineChange(text string, changed_line_index int, focused_line int, focused_col int) {
	log.Printf("saving %s at %d", text, stack.curr_index)
	stack.changes = append(stack.changes, newLineChange(text, changed_line_index, focused_line, focused_col))
	stack.curr_index++
}

func (stack *changeStack) commitLineAdd(text string, changed_line_index int, focused_line int, focused_col int) {
	stack.changes = append(stack.changes, newLineAdd(text, changed_line_index, focused_line, focused_col))
	stack.curr_index++
}

func (stack *changeStack) commitLineRemove(text string, changed_line_index int, focused_line int, focused_col int) {
	stack.changes = append(stack.changes, newLineRemove(text, changed_line_index, focused_line, focused_col))
	stack.curr_index++
}

func (stack *changeStack) undoChange(content []string) {
	if stack.curr_index > 0 {
		stack.curr_index--
		switch _change := stack.changes[stack.curr_index].(type) {
		case *textChange:
			log.Printf("undoing %s = %s at %d", content[_change.line], _change.original_text, stack.curr_index)
			content[_change.line] = _change.original_text
		}
	}
}

func (stack *changeStack) redoChange(content []string) {
	if stack.curr_index < len(stack.changes) {
		switch _change := stack.changes[stack.curr_index].(type) {
		case *textChange:
			content[_change.line] = _change.original_text
		}
		stack.curr_index++
	}
}
