package containeragentuc

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/containeragentuc/containeragentdto"
)

// fakeStdin is a fake exec's input: what was written to it, and whether it was
// closed.
type fakeStdin struct {
	bytes.Buffer
	closed bool
}

func (s *fakeStdin) CloseWrite() error {
	s.closed = true
	return nil
}

func inputs(msgs []*containeragentdto.ExecInput, end error) func() (*containeragentdto.ExecInput, error) {
	return func() (*containeragentdto.ExecInput, error) {
		if len(msgs) == 0 {
			return nil, end
		}
		msg := msgs[0]
		msgs = msgs[1:]
		return msg, nil
	}
}

// When the client has sent all of the stdin, the command is told its input
// ended: a command reading to the end, such as psql loading a dump, then
// finishes rather than waits forever.
func TestForwardExecInputClosesStdinWhenTheClientIsDone(t *testing.T) {
	stdin := &fakeStdin{}
	var resized []uint32

	forwardExecInput(inputs([]*containeragentdto.ExecInput{
		{Stdin: []byte("CREATE ")},
		{Resize: &containeragentdto.ResizeOptions{Width: 80, Height: 24}},
		{Stdin: []byte("TABLE t();\n")},
	}, io.EOF), stdin, func(w, h uint32) { resized = append(resized, w, h) })

	assert.Equal(t, "CREATE TABLE t();\n", stdin.String())
	assert.Equal(t, []uint32{80, 24}, resized)
	assert.True(t, stdin.closed)
}

// A stream that breaks is not an end of input: the exec goes with the stream.
func TestForwardExecInputLeavesStdinOpenWhenTheStreamBreaks(t *testing.T) {
	stdin := &fakeStdin{}

	forwardExecInput(inputs(nil, errors.New("transport is closing")), stdin, func(uint32, uint32) {})

	assert.False(t, stdin.closed)
}
