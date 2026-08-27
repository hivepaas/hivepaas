package nodeagentuc

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging/mocks"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/nodeexecservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/nodeagentuc/nodeagentdto"
)

func TestExecuteCommand(t *testing.T) {
	logger := &mocks.Logger{}
	uc := New(logger, nil, nil)

	t.Run("empty command returns error", func(t *testing.T) {
		_, err := uc.ExecuteCommand(context.Background(), &nodeagentdto.ExecCommandReq{
			CommandExecOpts: nil,
		})
		assert.Error(t, err)
	})

	t.Run("successful command execution streams stdout to writer", func(t *testing.T) {
		var stdoutBuf bytes.Buffer
		resp, err := uc.ExecuteCommand(context.Background(), &nodeagentdto.ExecCommandReq{
			CommandExecOpts: &nodeexecservice.CommandExecOpts{
				Command: []string{"echo", "hello hivepaas"},
				Stdout:  &stdoutBuf,
			},
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, int32(0), resp.ExitCode)
		assert.Contains(t, stdoutBuf.String(), "hello hivepaas")
	})

	t.Run("command with custom environment variables", func(t *testing.T) {
		var stdoutBuf bytes.Buffer
		resp, err := uc.ExecuteCommand(context.Background(), &nodeagentdto.ExecCommandReq{
			CommandExecOpts: &nodeexecservice.CommandExecOpts{
				Command: []string{"sh", "-c", "echo $TEST_VAR"},
				Env:     []string{"TEST_VAR=custom_value"},
				Stdout:  &stdoutBuf,
			},
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, int32(0), resp.ExitCode)
		assert.Contains(t, stdoutBuf.String(), "custom_value")
	})

	t.Run("non-zero exit code captures stderr to writer", func(t *testing.T) {
		var stderrBuf bytes.Buffer
		resp, err := uc.ExecuteCommand(context.Background(), &nodeagentdto.ExecCommandReq{
			CommandExecOpts: &nodeexecservice.CommandExecOpts{
				Command: []string{"sh", "-c", "echo 'error message' >&2; exit 42"},
				Stderr:  &stderrBuf,
			},
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, int32(42), resp.ExitCode)
		assert.Contains(t, stderrBuf.String(), "error message")
	})

	t.Run("merged stderr into stdout when only stdout writer is provided", func(t *testing.T) {
		var combinedBuf bytes.Buffer
		resp, err := uc.ExecuteCommand(context.Background(), &nodeagentdto.ExecCommandReq{
			CommandExecOpts: &nodeexecservice.CommandExecOpts{
				Command: []string{"sh", "-c", "echo 'normal out'; echo 'error out' >&2"},
				Stdout:  &combinedBuf,
			},
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, int32(0), resp.ExitCode)
		assert.Contains(t, combinedBuf.String(), "normal out")
		assert.Contains(t, combinedBuf.String(), "error out")
	})

	t.Run("remote interrupt via context cancellation kills process group", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		// Cancel after 100ms
		go func() {
			time.Sleep(100 * time.Millisecond)
			cancel()
		}()

		start := time.Now()
		resp, err := uc.ExecuteCommand(ctx, &nodeagentdto.ExecCommandReq{
			CommandExecOpts: &nodeexecservice.CommandExecOpts{
				Command: []string{"sleep", "10"},
			},
		})

		duration := time.Since(start)
		assert.Error(t, err)
		// Process should have been terminated quickly upon cancel, well before 10s
		assert.Less(t, duration, 5*time.Second)
		assert.NotNil(t, resp)
	})
}
