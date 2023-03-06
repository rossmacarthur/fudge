package errors

import (
	"fmt"
)

// Error is a concrete error type containing a stack trace
type Error struct {
	// message is the optional sentinel message (can be empty)
	message string
	// code is the optional sentinel error code (can be empty)
	code string
	// original is the original non-Fudge error (can be nil)
	original error
	// trace is the stack trace
	// contextual messages and key values are attached to individual stack frames
	trace []Frame
}

// Clone deep copies the error
func (e *Error) Clone() *Error {
	c := *e
	c.trace = make([]Frame, 0, len(e.trace))
	for _, f := range e.trace {
		c.trace = append(c.trace, *f.Clone())
	}
	return &c
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
	if e.message != "" {
		fmt.Fprintf(s, "%s (%s)\n", e.message, e.code)
	}
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
