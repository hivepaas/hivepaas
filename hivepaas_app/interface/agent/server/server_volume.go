package server

import (
	"context"

	agentproto "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/proto"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/server/volumeservice"
)

func (s *AgentServer) RemoveVolume(
	ctx context.Context,
	req *agentproto.RemoveVolumeReq,
) (*agentproto.RemoveVolumeResp, error) {
	return volumeservice.RemoveVolume(ctx, s.volumeAgentUC, req) //nolint:wrapcheck
}
