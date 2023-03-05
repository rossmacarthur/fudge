package errors

import (
	"fmt"
)

// Error is a concrete error type containing a stack trace
type Error struct {
	stack []Frame
}

func (e *Error) Error() string {
	return e.String()
}

func (e *Error) String() string {
	return fmt.Sprintf("%v", e)
}

func (e Error) Format(s fmt.State, verb rune) {
	for i, f := range e.stack {
		if i > 0 {
			fmt.Fprint(s, "\n")
		}
		f.Format(s, verb)
	}
}
