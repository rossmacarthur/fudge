package errors

import (
	"errors"

	"github.com/rossmacarthur/fudge"
)

// New creates a new error with a message and options.
//
// This method is intended to be used to define local errors. A stack trace is
// is attached to the error.
func New(msg string, opts ...fudge.Option) error {
	errors := Error{trace: trace(1)}
	frame := &errors.trace[0]
	frame.message = msg
	for _, o := range opts {
		o.Apply(frame)
	}
	return &errors
}

// NewSentinel creates a new sentinel error with a message and code.
//
// This method is intended to be used to define global sentinel errors. No
// stack trace is attached until these errors are wrapped.
func NewSentinel(msg string, code string) error {
	return &Error{message: msg, code: code}
}

// Wrap wraps an existing error with a new message and options.
//
// This method consumes the original error, so it should not be used afterward.
func Wrap(err error, msg string, opts ...fudge.Option) error {
	if err == nil {
		return nil
	}

	errors, ok := err.(*Error)
	if ok && errors.trace == nil {
		// wrapping a sentinel Fudge error
		errors = errors.Clone()
		errors.trace = trace(1)
		frame := &errors.trace[0]
		frame.message = msg
		for _, o := range opts {
			o.Apply(frame)
		}

	} else if ok {
		// wrapping a Fudge error
		errors = errors.Clone()
		frame := findCallSite(errors, 1)
		frame.message = msg
		for _, o := range opts {
			o.Apply(frame)
		}

	} else {
		// wrapping a non-Fudge error
		errors = &Error{original: err, trace: trace(1)}
		frame := &errors.trace[0]
		frame.message = msg
		for _, o := range opts {
			o.Apply(frame)
		}
	}

	return errors
}

func findCallSite(e *Error, skip int) *Frame {
	file, line := call(skip + 1)
	for i, f := range e.trace {
		if f.file == file && f.line == line {
			return &e.trace[i]
		}
	}

	// the call site doesn't exist in the trace so we need to add it
	// by combining the traces
	trace := trace(skip + 1)
outer:
	for _, f := range trace {
		for j, g := range e.trace {
			if f.file == g.file && f.line == g.line {
				e.trace = append(e.trace[:j], trace...)
				break outer
			}
		}
	}

	for i, f := range e.trace {
		if f.file == file && f.line == line {
			return &e.trace[i]
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
