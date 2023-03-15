package errors

import (
	"fmt"

	"github.com/rossmacarthur/fudge/internal/stack"
)

// Frame is a single frame in a stack trace
type Frame struct {
	// File is the file name associated with the frame
	File string
	// function is the function name associated with the frame
	Function string
	// Line is the fine number associated with the frame
	Line int
	// Message is the message associated with the frame (can be empty)
	Message string
	// KeyValues is a map of key-value pairs associated with the frame (can be nil)
	KeyValues KeyValues
}

func (f *Frame) clone() *Frame {
	c := *f
	if f.KeyValues != nil {
		c.KeyValues = f.KeyValues.clone()
	}
	return &c
}

func isGlobal(t []Frame) bool {
	for _, f := range t {
		if f.File == "runtime/proc.go" && f.Function == "doInit" {
			return true
		}
	}
	return false
}

func (f Frame) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v', 's':
		fmt.Fprintf(s, "%s:%d %s", f.File, f.Line, f.Function)
	default:
		fmt.Fprintf(s, "%%!%c(Frame=%s:%d)", verb, f.File, f.Line)
	}
}

func trace(skip int) []Frame {
	var trace []Frame
	for _, f := range stack.Trace(skip + 1) {
		trace = append(trace, Frame{
			File:     f.File,
			Function: f.Function,
			Line:     f.Line,
		})
	}
	return trace
}

func call(skip int) *Frame {
	f := stack.Call(skip + 1)
	return &Frame{
		File:     f.File,
		Function: f.Function,
		Line:     f.Line,
	}
}
