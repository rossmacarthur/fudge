package errors

import (
	"fmt"
)

// Error is a concrete error type containing a stack trace
type Error struct {
	// original is the original non-Fudge error (can be nil)
	original error
	// trace is the stack trace
	// contextual messages and key values are attached to individual stack frames
	trace []Frame
}

// Unwrap implements the errors.Unwrap interface
func (e *Error) Unwrap() error {
	return e.original
}

func (e *Error) Error() string {
	return e.String()
}

func (e *Error) String() string {
	return fmt.Sprintf("%v", e)
}

func (e Error) Format(s fmt.State, verb rune) {
	if e.original != nil {
		fmt.Fprintf(s, "%s\n", e.original.Error())
	}
	for i, f := range e.trace {
		if i > 0 {
			fmt.Fprint(s, "\n")
		}
		f.Format(s, verb)
	}
}
