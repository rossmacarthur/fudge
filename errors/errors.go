package errors

import (
	"errors"

	"github.com/rossmacarthur/fudge"
)

// New creates a new error with a message and options.
func New(msg string, opts ...fudge.Option) error {
	errors := Error{stack: trace(2)}
	frame := &errors.stack[0]
	frame.message = msg
	for _, o := range opts {
		o.Apply(frame)
	}
	return &errors
}

// Wrap wraps an existing error with a new message and options.
//
// This method consumes the original error, so it should not be used afterward.
func Wrap(err error, msg string, opts ...fudge.Option) error {
	if err == nil {
		return nil
	}

	errors, ok := err.(*Error)
	if ok {
		frame := findCallSite(errors)
		frame.message = msg
		for _, o := range opts {
			o.Apply(frame)
		}
	} else {
		errors = &Error{stack: trace(2)}
		frame := &errors.stack[0]
		frame.original = err
		frame.message = msg
		for _, o := range opts {
			o.Apply(frame)
		}
	}

	return errors
}

func findCallSite(e *Error) *Frame {
	const skip int = 3

	file, line := call(skip)
	for i, f := range e.stack {
		if f.file == file && f.line == line {
			return &e.stack[i]
		}
	}

	// the call site doesn't exist in the stack so we need to add it
	// by combining the traces
	stack := trace(skip)
outer:
	for _, f := range stack {
		for j, g := range e.stack {
			if f.file == g.file && f.line == g.line {
				e.stack = append(e.stack[:j], stack...)
				break outer
			}
		}
	}

	for i, f := range e.stack {
		if f.file == file && f.line == line {
			return &e.stack[i]
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
