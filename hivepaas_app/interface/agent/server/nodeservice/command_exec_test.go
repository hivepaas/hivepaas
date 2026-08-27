package nodeservice

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	agentproto "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/proto"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging/mocks"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/nodeagentuc"
)

func TestExecuteCommand(t *testing.T) {
	logger := &mocks.Logger{}
	uc := nodeagentuc.New(logger, nil, nil)

	t.Run("nil request returns empty response", func(t *testing.T) {
		resp, err := ExecuteCommand(context.Background(), uc, nil)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("valid command returns output and exit code", func(t *testing.T) {
		req := &agentproto.ExecCommandReq{
			Command: []string{"echo", "hello agent grpc"},
		}
		resp, err := ExecuteCommand(context.Background(), uc, req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, int32(0), resp.GetExitCode())
		assert.Contains(t, string(resp.GetStdout()), "hello agent grpc")
	})
}
