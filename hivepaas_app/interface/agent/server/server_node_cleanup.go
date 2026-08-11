package server

import (
	"context"

	agentproto "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/proto"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/server/nodecleanupservice"
)

func (s *AgentServer) NodeCleanup(
	ctx context.Context,
	req *agentproto.NodeCleanupReq,
) (*agentproto.NodeCleanupResp, error) {
	return nodecleanupservice.NodeCleanup(ctx, s.nodeCleanupAgentUC, req) //nolint:wrapcheck
}
