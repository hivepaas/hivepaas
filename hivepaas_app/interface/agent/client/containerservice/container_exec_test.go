package containerservice

import (
	"bytes"
	"context"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/moby/moby/api/pkg/stdcopy"
	"github.com/stretchr/testify/assert"
)

func newTestStream(muxOutput bool) *ContainerExecStream {
	s := &ContainerExecStream{muxOutput: muxOutput}
	s.cond = sync.NewCond(&s.readMutex)
	return s
}

// A command run without a TTY reads as Docker's own attach stream does: each
// chunk framed with its stream, so stdout and stderr are told apart again.
func TestNonTTYOutputIsFramedAsDockerFramesIt(t *testing.T) {
	s := newTestStream(true)
	s.handleResp(&ContainerExecResp{Stdout: []byte("hello\n")})
	s.handleResp(&ContainerExecResp{Stderr: []byte("oops\n")})
	s.handleResp(&ContainerExecResp{Stdout: []byte("bye\n")})
	s.setReadErr(io.EOF)

	var stdout, stderr bytes.Buffer
	_, err := stdcopy.StdCopy(&stdout, &stderr, s.ToExecAttachResult().Reader)

	assert.NoError(t, err)
	assert.Equal(t, "hello\nbye\n", stdout.String())
	assert.Equal(t, "oops\n", stderr.String())
}

// With a TTY the output is one raw stream, as Docker gives it.
func TestTTYOutputIsRaw(t *testing.T) {
	s := newTestStream(false)
	s.handleResp(&ContainerExecResp{Stdout: []byte("hello\n")})
	s.setReadErr(io.EOF)

	out, err := io.ReadAll(s.ToExecAttachResult().Reader)

	assert.NoError(t, err)
	assert.Equal(t, "hello\n", string(out))
}

// The exit code comes last on the stream: asking for it waits until the agent
// has sent it, or the stream has ended without it.
func TestWaitExitCodeWaitsForTheAgentToSendIt(t *testing.T) {
	s := newTestStream(true)
	go func() {
		time.Sleep(50 * time.Millisecond)
		code := int32(3)
		s.handleResp(&ContainerExecResp{ExitCode: &code})
	}()

	code, ok := s.WaitExitCode(context.Background())

	assert.True(t, ok)
	assert.Equal(t, int32(3), code)
}

func TestWaitExitCodeEndsWithTheStream(t *testing.T) {
	s := newTestStream(true)
	go func() {
		time.Sleep(50 * time.Millisecond)
		s.setReadErr(io.EOF)
	}()

	_, ok := s.WaitExitCode(context.Background())

	assert.False(t, ok)
}

func TestWaitExitCodeEndsWithTheContext(t *testing.T) {
	s := newTestStream(true)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, ok := s.WaitExitCode(ctx)

	assert.False(t, ok)
}
