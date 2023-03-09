package errors

import (
	"fmt"
	"runtime"
	"strings"
)

// Frame is a single frame in a stack trace
type Frame struct {
	// Message is the message associated with the frame (can be empty)
	Message string
	// KeyValues is a map of key-value pairs associated with the frame (can be nil)
	KeyValues KeyValues
	// File is the file name associated with the frame
	File string
	// Line is the fine number associated with the frame
	Line int
}

func (f *Frame) clone() *Frame {
	c := *f
	if f.KeyValues != nil {
		c.KeyValues = f.KeyValues.clone()
	}
	return &c
}

// SetKeyValue implements the fudge.apply interface
func (f *Frame) SetKeyValue(k, v string) {
	if f.KeyValues == nil {
		f.KeyValues = make(KeyValues)
	}
	f.KeyValues[k] = v
}

func (f Frame) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v', 's':
		fmt.Fprintf(s, "%s:%d", f.File, f.Line)
		if f.Message != "" {
			fmt.Fprintf(s, ": %s", f.Message)
		}
		if len(f.KeyValues) > 0 && s.Flag(int('#')) {
			fmt.Fprintf(s, " {%v}", f.KeyValues)
		}
	default:
		fmt.Fprintf(s, "%%!%c(Frame=%s:%d)", verb, f.File, f.Line)
	}
}

func trace(skip int) []Frame {
	var pcs [512]uintptr
	n := runtime.Callers(skip+1, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])
	trace := make([]Frame, 0, n)
	frame, more := frames.Next()
	for more {
		frame, more = frames.Next()
		trace = append(trace, Frame{
			File: pkgFilePath(frame.Function, frame.File),
			Line: frame.Line,
		})
	}
	return trace
}

func call(skip int) (string, int) {
	var pcs [3]uintptr
	n := runtime.Callers(skip+1, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])
	frame, _ := frames.Next()
	frame, _ = frames.Next()
	return pkgFilePath(frame.Function, frame.File), frame.Line
}

func pkgFilePath(function, file string) string {
	const pathSep = "/"

	var pre, post string

	lastSep := strings.LastIndex(file, pathSep)
	if lastSep == -1 {
		post = file
	} else {
		post = file[strings.LastIndex(file[:lastSep], pathSep)+1:]
	}

	end := strings.LastIndex(function, pathSep)
	if end == -1 {
		return post
	}
	pre = function[:end]

	return pre + pathSep + post
}
