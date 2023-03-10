package errors

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Error is a concrete error type containing a stack trace
type Error struct {
	// Binary is the name of the executable the error occurred in
	Binary string
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
		fmt.Fprintf(s, "%s", e.fullMessage())

		switch {
		case s.Flag(int('+')):
			for _, f := range e.Trace {
				fmt.Fprint(s, "\n")
				f.Format(s, verb)
			}
		case s.Flag(int('#')):
			kvs := e.fullKeyValues()
			if len(kvs) > 0 {
				fmt.Fprintf(s, " {%v}", kvs)
			}
			for _, f := range e.Trace {
				fmt.Fprint(s, "\n")
				f.Format(s, verb)
			}
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
		if e.Code != "" {
			s.WriteString(" (")
			s.WriteString(e.Code)
			s.WriteString(")")
		}
	}
	if e.Original != nil {
		write(e.Original.Error())
	}

	return s.String()
}

// fullKeyValues returns the full key values
func (e *Error) fullKeyValues() KeyValues {
	kvs := make(KeyValues)
	for i := len(e.Trace) - 1; i >= 0; i-- {
		for k, v := range e.Trace[i].KeyValues {
			kvs[k] = v
		}
	}
	return kvs
}

func binary() string {
	return filepath.Base(os.Args[0])
}
