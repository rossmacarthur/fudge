package stack

import (
	"runtime"
	"strings"
)

// Frame is a simplified version of runtime.Frame
type Frame struct {
	File     string
	Function string
	Line     int
}

func Trace(skip int) []Frame {
	var pcs [512]uintptr
	n := runtime.Callers(skip+1, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])
	trace := make([]Frame, 0, n)
	frame, more := frames.Next()
	for more {
		frame, more = frames.Next()
		trace = append(trace, Frame{
			File:     tidyFile(frame.Function, frame.File),
			Function: tidyFunction(frame.Function),
			Line:     frame.Line,
		})
	}
	return trace
}

func Call(skip int) Frame {
	var pcs [3]uintptr
	n := runtime.Callers(skip+1, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])
	frame, _ := frames.Next()
	frame, _ = frames.Next()
	return Frame{
		File:     tidyFile(frame.Function, frame.File),
		Function: tidyFunction(frame.Function),
		Line:     frame.Line,
	}
}

const pathSep = "/"
const pkgSep = "."

func tidyFunction(function string) string {
	if i := strings.LastIndex(function, pathSep); i != -1 {
		function = function[i+len(pathSep):]
	}
	if i := strings.Index(function, pkgSep); i != -1 {
		function = function[i+len(pkgSep):]
	}
	return function
}

func tidyFile(function, file string) string {
	var pre, post string

	i := strings.LastIndex(file, pathSep)
	if i == -1 {
		post = file
	} else {
		post = file[strings.LastIndex(file[:i], pathSep)+1:]
	}

	i = strings.LastIndex(function, pathSep)
	if i == -1 {
		return post
	}
	pre = function[:i]

	return pre + pathSep + post
}
