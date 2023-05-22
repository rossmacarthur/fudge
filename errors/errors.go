package errors

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rossmacarthur/fudge"
)

// Sentinel creates a new sentinel error with a message and code.
//
// This method is intended to be used to define global sentinel errors. These
// errors can be used with Is to check for equality even over gRPC (if the
// provided gRPC interceptors are used). No stack trace is attached and these
// errors must be wrapped with Wrap when they are used in order to attach one.
func Sentinel(msg string, code string) error {
	return &Error{Message: msg, Code: code}
}

// New creates a new error with a message and options.
//
// These errors can be used with Is to check for equality but not over gRPC. If
// that is required then Sentinel should be used. If used in the global scope
// then no stack trace is attached and these errors must be wrapped with Wrap
// when they are used in order to attach one.
func New(msg string, opts ...fudge.Option) error {
	t := trace(1)
	if isGlobal(t) {
		if len(opts) > 0 {
			panic("fudge/errors: options not allowed in conjunction with global errors")
		}
		return &Error{Message: msg}
	}

	errors := &Error{Binary: binary(), Trace: t}
	frame := &errors.Trace[0]
	frame.Message = msg
	applyOptions(frame, opts)
	return errors
}

// NewWithCause creates a new error with a message, cause and options.
//
// This method is intended to be used when the cause is a Fudge error and you
// don't want to use Wrap which merges the stack trace. Most of the time you
// want to use Wrap.
func NewWithCause(msg string, cause error, opts ...fudge.Option) error {
	errors := &Error{Binary: binary(), Cause: cause, Trace: trace(1)}
	frame := &errors.Trace[0]
	frame.Message = msg
	applyOptions(frame, opts)
	return errors
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
		applyOptions(frame, opts)

	} else if ok {
		// wrapping a Fudge error
		errors = errors.clone()
		frame := findCallSite(errors, 1)
		if frame.Message == "" {
			frame.Message = msg
		} else {
			frame.Message = fmt.Sprintf("%s: %s", msg, frame.Message)
		}
		applyOptions(frame, opts)

	} else {
		// wrapping a non-Fudge error
		errors = &Error{Binary: binary(), Cause: err, Trace: trace(1)}
		frame := &errors.Trace[0]
		frame.Message = msg
		applyOptions(frame, opts)
	}

	return errors
}

func binary() string {
	return filepath.Base(os.Args[0])
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
