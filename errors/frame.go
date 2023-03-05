package errors

import (
	"fmt"
	"strings"

	"github.com/go-stack/stack"
)

// Frame is a single frame in a stack trace
type Frame struct {
	// original is the original non-Fudge error (can be nil)
	original error
	// message is the message associated with the frame (can be empty)
	message string
	// keyValues is a map of key-value pairs associated with the frame (can be nil)
	keyValues KeyValues
	// file is the file name of the frame
	file string
	// line is the line number of the frame
	line int
}

// SetKeyValue implements the fudge.apply interface
func (f *Frame) SetKeyValue(k, v string) {
	if f.keyValues == nil {
		f.keyValues = make(map[string]string)
	}
	f.keyValues[k] = v
}

func (f *Frame) String() string {
	return fmt.Sprintf("%v", f)
}

func (f Frame) Format(s fmt.State, verb rune) {
	fmt.Fprintf(s, "%s:%d", f.file, f.line)
	if f.message != "" {
		fmt.Fprintf(s, ": %s", f.message)
	}
	if f.original != nil {
		fmt.Fprintf(s, ": %s", f.original.Error())
	}
	if len(f.keyValues) > 0 && (s.Flag(int('#')) || s.Flag(int('+'))) {
		fmt.Fprintf(s, " {%v}", f.keyValues)
	}
}

func trace(skip int) []Frame {
	var trace []Frame
	for _, c := range stack.Trace()[skip:] {
		f := c.Frame()
		trace = append(trace, Frame{
			file: pkgFilePath(f.Function, f.File),
			line: f.Line,
		})
	}

	return trace
}

func call(skip int) (string, int) {
	c := stack.Caller(skip)
	f := c.Frame()
	return pkgFilePath(f.Function, f.File), f.Line
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
