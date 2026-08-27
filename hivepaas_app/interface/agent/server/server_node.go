package server

import (
	"context"

	agentproto "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/proto"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/server/nodeservice"
)

func (s *AgentServer) ExecuteCommand(
	ctx context.Context,
	req *agentproto.ExecCommandReq,
) (*agentproto.ExecCommandResp, error) {
	return nodeservice.ExecuteCommand(ctx, s.nodeAgentUC, req) //nolint:wrapcheck
}
