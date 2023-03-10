package errors

import (
	"errors"
	"fmt"

	"github.com/rossmacarthur/fudge"
)

// New creates a new error with a message and options.
//
// If used in the global scope no stack trace will be attached until the error
// is wrapped with Wrap.
func New(msg string, opts ...fudge.Option) error {
	trace := trace(1)

	c := call(2)
	if c.File == "runtime/proc.go" && c.Function == "doInit" {
		return &Error{Message: msg}
	}

	errors := Error{Trace: trace}
	frame := &errors.Trace[0]
	frame.Message = msg
	for _, o := range opts {
		o.Apply(frame)
	}
	return &errors
}

// NewSentinel creates a new sentinel error with a message and code.
//
// This method is intended to be used to define global sentinel errors. No
// stack trace is attached until these errors are wrapped with Wrap.
func NewSentinel(msg string, code string) error {
	return &Error{Message: msg, Code: code}
}

// Wrap wraps an existing error with a new message and options and the
// underlying cause of the error will remain unchanged.
//
// If the error is a
//   - sentinel Fudge error then it is cloned and a stack trace is added.
//   - inline Fudge error then the stack trace is extended and/or annotated
//   - non-Fudge error it is converted to a Fudge error and a trace back is
//     added and the original error is available via the Unwrap method.
func Wrap(err error, msg string, opts ...fudge.Option) error {
	if err == nil {
		return nil
	}

	errors, ok := err.(*Error)
	if ok && errors.Trace == nil {
		// wrapping a sentinel Fudge error
		errors = errors.clone()
		errors.Trace = trace(1)
		frame := &errors.Trace[0]
		frame.Message = msg
		for _, o := range opts {
			o.Apply(frame)
		}

	} else if ok {
		// wrapping a Fudge error
		errors = errors.clone()
		frame := findCallSite(errors, 1)
		if frame.Message == "" {
			frame.Message = msg
		} else {
			frame.Message = fmt.Sprintf("%s: %s", msg, frame.Message)
		}
		for _, o := range opts {
			o.Apply(frame)
		}

	} else {
		// wrapping a non-Fudge error
		errors = &Error{Original: err, Trace: trace(1)}
		frame := &errors.Trace[0]
		frame.Message = msg
		for _, o := range opts {
			o.Apply(frame)
		}
	}

	return errors
}

func findCallSite(e *Error, skip int) *Frame {
	c := call(skip + 1)

	for i, f := range e.Trace {
		if f.File == c.File && f.Function == c.Function && f.Line == c.Line {
			return &e.Trace[i]
		}
	}

	// the call site doesn't exist in the trace so we need to add it
	// by combining the traces
	trace := trace(skip + 1)
outer:
	for _, f := range trace {
		for j, g := range e.Trace {
			if f.File == g.File && f.Function == g.Function && f.Line == g.Line {
				e.Trace = append(e.Trace[:j], trace...)
				break outer
			}
		}
	}

	for i, f := range e.Trace {
		if f.File == c.File && f.Function == c.Function && f.Line == c.Line {
			return &e.Trace[i]
		}
	}

	panic("failed to find call site frame")
}

// Unwrap is the same as the standard library's errors.Unwrap.
func Unwrap(err error) error {
	return errors.Unwrap(err)
}

// Is is the same as the standard library's errors.Is.
func Is(err error, target error) bool {
	return errors.Is(err, target)
}

// As is the same as the standard library's errors.As.
func As(err error, target any) bool {
	return errors.As(err, target)
}
