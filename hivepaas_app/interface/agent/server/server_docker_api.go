package server

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	agentproto "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/proto"
)

func (s *AgentServer) SyncDockerAPI(
	ctx context.Context,
	_ *agentproto.DockerAPISyncReq,
) (*agentproto.DockerAPISyncResp, error) {
	apps, err := s.dockerAPIAgentUC.Sync(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &agentproto.DockerAPISyncResp{Apps: int32(apps)}, nil //nolint:gosec // a count of apps
}

func (s *AgentServer) RemoveDockerAPIApp(
	ctx context.Context,
	req *agentproto.DockerAPIRemoveAppReq,
) (*agentproto.DockerAPIRemoveAppResp, error) {
	removed, err := s.dockerAPIAgentUC.RemoveApp(ctx, req.GetAppId())
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	//nolint:gosec // counts of what one node held
	return &agentproto.DockerAPIRemoveAppResp{
		Containers: int32(removed.Containers),
		Networks:   int32(removed.Networks),
		Volumes:    int32(removed.Volumes),
	}, nil
}
