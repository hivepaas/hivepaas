package errstack

import (
	"errors"
	"fmt"
	"io/fs"
	"testing"

	"github.com/stretchr/testify/assert"
)

func arise() error { return Wrap(errors.New("something broke"), 0) }

// The trace starts where the error arose, names that function even when it is
// inlined into its caller, and is headed by the error's type and message.
func TestATraceStartsWhereTheErrorArose(t *testing.T) {
	trace := Trace(arise())

	assert.Contains(t, trace, "*errors.errorString something broke\n")
	assert.Contains(t, trace, "errstack.arise")
	assert.Contains(t, trace, "errstack.TestATraceStartsWhereTheErrorArose")
	assert.NotContains(t, trace, "errstack.Wrap", "the trace starts at Wrap's caller")
}

// skip moves the start up the stack, as a helper wrapping for its caller does.
func TestSkipStartsTheTraceAtAnOuterCaller(t *testing.T) {
	helper := func(err error) error { return Wrap(err, 1) }

	trace := Trace(helper(errors.New("x")))

	assert.Contains(t, trace, "errstack.TestSkipStartsTheTraceAtAnOuterCaller")
	assert.NotContains(t, trace, "func1", "the helper itself is skipped")
}

// An error that carries a stack keeps it: wrapping it again does not move it.
func TestWrappingAgainKeepsTheFirstStack(t *testing.T) {
	err := arise()

	assert.Same(t, err, Wrap(err, 0))
	assert.Equal(t, Trace(err), Trace(Wrap(err, 0)))
}

// The message, errors.Is and errors.As go through to the error inside.
func TestTheErrorInsideIsReachable(t *testing.T) {
	err := Wrap(fmt.Errorf("reading: %w", fs.ErrNotExist), 0)

	assert.Equal(t, "reading: file does not exist", err.Error())
	assert.ErrorIs(t, err, fs.ErrNotExist)
	var pathErr *fs.PathError
	assert.False(t, errors.As(err, &pathErr))
}

// The trace is found through context added around the error, and the outermost
// stack is the one read.
func TestATraceIsFoundThroughAddedContext(t *testing.T) {
	err := fmt.Errorf("listing: %w", arise())

	assert.Contains(t, Trace(err), "errstack.arise")
	assert.Empty(t, Trace(errors.New("no stack")))
	assert.Empty(t, Trace(nil))
	assert.NoError(t, Wrap(nil, 0))
}
