package nodeexecserviceimpl

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/service/nodeexecservice"
)

type mockAgentService struct {
	addrByID    map[string]string
	addrByLabel map[string]string
}

func (m *mockAgentService) GetAgentAddrForNode(ctx context.Context, nodeID string) (string, error) {
	if addr, ok := m.addrByID[nodeID]; ok {
		return addr, nil
	}
	return "", assert.AnError
}

func (m *mockAgentService) GetAgentAddrForNodeLabel(ctx context.Context, nodeLabel string) (string, error) {
	if addr, ok := m.addrByLabel[nodeLabel]; ok {
		return addr, nil
	}
	return "", assert.AnError
}

func TestExecCommand_AgentAddressError(t *testing.T) {
	mockAgent := &mockAgentService{
		addrByID:    make(map[string]string),
		addrByLabel: make(map[string]string),
	}

	svc := &service{
		agentService: mockAgent,
	}

	var stdoutBuf bytes.Buffer
	_, err := svc.ExecCommand(context.Background(), &nodeexecservice.CommandExecReq{
		NodeID: "unknown-node",
		CommandExecOpts: &nodeexecservice.CommandExecOpts{
			Command: []string{"echo", "hi"},
			Stdout:  &stdoutBuf,
		},
	})
	assert.Error(t, err)
}
