package errors

import (
	"fmt"
	"strings"
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
	//
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

// Is implements the errors.Is interface
//
// A Fudge error is the same as another if they have the same error code. This
// means that you can only compare to a sentinel error.
func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	if !ok {
		return false // errors.Is will Unwrap and compare to original
	}

	if e.code != "" {
		return e.code == t.code
	}

	return e == t
}

// Unwrap implements the errors.Unwrap interface
func (e *Error) Unwrap() error {
	return e.original
}

func (e *Error) Error() string {
	return e.String()
}

func (e *Error) String() string {
	return fmt.Sprintf("%s", e)
}

// Format implements the fmt.Formatter interface
//
// The following verbs are supported:
//
//	%v, %s: print the wrapping messages and error message
//	%+v, %+s: print the error message and stack trace with wrapping messages
//	%#v, %#s: print the error message and stack trace with wrapping messages and key values
func (e *Error) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v', 's':
		if s.Flag(int('+')) || s.Flag(int('#')) {
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
				f.Format(s, verb) // Note: behaviour is different for + and # flags
			}
		} else {
			fmt.Fprintf(s, "%s", e.Message())
		}
	default:
		fmt.Fprintf(s, "%%!%c(*errors.Error=%s)", verb, e.Message())
	}
}

// Message returns the full error message
func (e *Error) Message() string {
	var s strings.Builder

	write := func(m string) {
		if s.Len() > 0 {
			s.WriteString(": ")
		}
		s.WriteString(m)
	}

	for i := len(e.trace) - 1; i >= 0; i-- {
		m := e.trace[i].message
		if m != "" {
			write(m)
		}
	}

	if e.message != "" {
		write(e.message)
		s.WriteString(" (")
		s.WriteString(e.code)
		s.WriteString(")")
	}
	if e.original != nil {
		write(e.original.Error())
	}

	return s.String()
}
