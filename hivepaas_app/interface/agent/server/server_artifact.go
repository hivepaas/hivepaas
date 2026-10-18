package server

import (
	"context"

	agentproto "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/proto"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/server/artifactservice"
)

func (s *AgentServer) Ensure(
	ctx context.Context,
	req *agentproto.EnsureArtifactReq,
) (*agentproto.EnsureArtifactResp, error) {
	return artifactservice.Ensure(ctx, s.artifactAgentUC, req) //nolint:wrapcheck
}
