package errors

import (
	"fmt"
	"strings"
)

// Error is a concrete error type containing a stack trace
type Error struct {
	// Message is the optional sentinel message (can be empty)
	Message string
	// Code is the optional sentinel error code (can be empty)
	Code string
	// Original is the original non-Fudge error (can be nil)
	Original error
	// Trace is the stack trace
	//
	// contextual messages and key values are attached to individual stack frames
	Trace []Frame
}

// Unwrap implements the errors.Unwrap interface and returns the original
// error if possible
func (e *Error) Unwrap() error {
	return e.Original
}

// clone deep copies the error
func (e *Error) clone() *Error {
	c := *e
	c.Trace = make([]Frame, 0, len(e.Trace))
	for _, f := range e.Trace {
		c.Trace = append(c.Trace, *f.clone())
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

	if e.Code != "" {
		return e.Code == t.Code
	}

	return e == t
}

func (e *Error) Error() string {
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
			if e.Message != "" {
				fmt.Fprintf(s, "%s (%s)", e.Message, e.Code)
				if len(e.Trace) > 0 {
					fmt.Fprint(s, "\n")
				}
			}
			if e.Original != nil {
				fmt.Fprintf(s, "%s", e.Original.Error())
				if len(e.Trace) > 0 {
					fmt.Fprint(s, "\n")
				}
			}
			for i, f := range e.Trace {
				if i > 0 {
					fmt.Fprint(s, "\n")
				}
				f.Format(s, verb) // Note: behaviour is different for + and # flags
			}
		} else {
			fmt.Fprintf(s, "%s", e.fullMessage())
		}
	default:
		fmt.Fprintf(s, "%%!%c(*errors.Error=%s)", verb, e.fullMessage())
	}
}

// fullMessage returns the full error message
func (e *Error) fullMessage() string {
	var s strings.Builder

	write := func(m string) {
		if s.Len() > 0 {
			s.WriteString(": ")
		}
		s.WriteString(m)
	}

	for i := len(e.Trace) - 1; i >= 0; i-- {
		m := e.Trace[i].Message
		if m != "" {
			write(m)
		}
	}

	if e.Message != "" {
		write(e.Message)
		s.WriteString(" (")
		s.WriteString(e.Code)
		s.WriteString(")")
	}
	if e.Original != nil {
		write(e.Original.Error())
	}

	return s.String()
}
