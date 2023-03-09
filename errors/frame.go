package errors

import (
	"fmt"
	"runtime"
	"strings"
)

// Frame is a single frame in a stack trace
type Frame struct {
	// message is the message associated with the frame (can be empty)
	message string
	// keyValues is a map of key-value pairs associated with the frame (can be nil)
	keyValues KeyValues
	// file is the file name of the frame
	file string
	// line is the line number of the frame
	line int
}

func (f *Frame) Clone() *Frame {
	c := *f
	if f.keyValues != nil {
		c.keyValues = f.keyValues.Clone()
	}
	return &c
}

// SetKeyValue implements the fudge.apply interface
func (f *Frame) SetKeyValue(k, v string) {
	if f.keyValues == nil {
		f.keyValues = make(KeyValues)
	}
	f.keyValues[k] = v
}

func (f *Frame) String() string {
	return fmt.Sprintf("%s", f)
}

func (f Frame) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v', 's':
		fmt.Fprintf(s, "%s:%d", f.file, f.line)
		if f.message != "" {
			fmt.Fprintf(s, ": %s", f.message)
		}
		if len(f.keyValues) > 0 && s.Flag(int('#')) {
			fmt.Fprintf(s, " {%v}", f.keyValues)
		}
	default:
		fmt.Fprintf(s, "%%!%c(Frame=%s:%d)", verb, f.file, f.line)
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
			file: pkgFilePath(frame.Function, frame.File),
			line: frame.Line,
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
