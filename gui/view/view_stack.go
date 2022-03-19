package view

type viewStack struct {
	arr []_viewName
	top int
}

func newStack() viewStack {
	return viewStack{
		arr: make([]_viewName, 0),
		top: 0,
	}
}

func (stack *viewStack) peek() _viewName {
	if stack.top == 0 {
		return none
	}
	return stack.arr[stack.top-1]
}

func (stack *viewStack) push(name _viewName) {
	stack.arr = append(stack.arr, name)
	stack.top++
}

func (stack *viewStack) pop() {
	if stack.top > 0 {
		stack.arr = stack.arr[:len(stack.arr)-1]
		stack.top--
	}
}
