package errors

import (
	"runtime"
	"strings"
)

// StackFrame describes one line in the stack trace.
type StackFrame struct {
	// File is the path to the file containing this ProgramCounter
	File string
	// LineNumber in that file
	LineNumber int
	// Name of the function that contains this ProgramCounter
	Name string
	// Package that contains this function
	Package string
	// ProgramCounter is the underlying pointer
	ProgramCounter uintptr
}

// NewStackFrame creates a stack frame object from a program counter.
func NewStackFrame(pc uintptr) StackFrame {
	frame := StackFrame{ProgramCounter: pc}

	// get the function for the program counter
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return frame
	}

	name := fn.Name()
	pkg := ""

	// program counters are return addresses, but we want the line that called us
	frame.File, frame.LineNumber = fn.FileLine(pc - 1)

	// clean the package names
	if lastslash := strings.LastIndex(name, "/"); lastslash >= 0 {
		pkg += name[:lastslash] + "/"
		name = name[lastslash+1:]
	}

	if period := strings.Index(name, "."); period >= 0 {
		pkg += name[:period]
		name = name[period+1:]
	}

	name = strings.Replace(name, "·", ".", -1)

	frame.Package = pkg
	frame.Name = name

	return frame
}
