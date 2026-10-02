// Package errstack keeps the stack an error arose on, and writes it out.
package errstack

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"strings"
)

// maxDepth is how many frames a stack keeps.
const maxDepth = 50

// stackError is an error and the stack it arose on.
type stackError struct {
	err error
	pcs []uintptr
}

func (e *stackError) Error() string { return e.err.Error() }

func (e *stackError) Unwrap() error { return e.err }

// Wrap returns err with the stack of Wrap's caller, or of a caller further out
// by skip. An error that already carries a stack is returned as it is: the
// stack of where it arose is the one worth keeping.
func Wrap(err error, skip int) error {
	if err == nil {
		return nil
	}
	if e, ok := err.(*stackError); ok { //nolint:errorlint // the error itself, not one it wraps
		return e
	}
	pcs := make([]uintptr, maxDepth)
	n := runtime.Callers(2+skip, pcs) //nolint:mnd // runtime.Callers and Wrap
	return &stackError{err: err, pcs: pcs[:n]}
}

// Trace is the stack of the outermost error carrying one in err's chain:
// headed by the type and message of the error it carries, then a frame per
// function, file and line first. It is empty when nothing in the chain has one.
func Trace(err error) string {
	e, ok := errors.AsType[*stackError](err)
	if !ok {
		return ""
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%s %s\n", reflect.TypeOf(e.err), e.err.Error())
	// runtime.CallersFrames, not runtime.FuncForPC of each pc: only it names a
	// function inlined into its caller, and reads a return address's line.
	frames := runtime.CallersFrames(e.pcs)
	for {
		frame, more := frames.Next()
		if frame.Function != "" {
			fmt.Fprintf(&b, "%s:%d\n\t%s\n", frame.File, frame.Line, frame.Function)
		}
		if !more {
			return b.String()
		}
	}
}
