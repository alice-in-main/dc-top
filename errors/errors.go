package errors

import "fmt"

type InvalidInput struct {
	Input string
}

func (e InvalidInput) Error() string {
	return fmt.Sprintf("Received invalid input: %s", e.Input)
}

func NewInvalidInputError(input string) InvalidInput {
	return InvalidInput{
		Input: input,
	}
}

type NotFoundError struct {
	Object string
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf("Didn't find: %s", e.Object)
}

func NewNotFoundError(object string) NotFoundError {
	return NotFoundError{
		Object: object,
	}
}
