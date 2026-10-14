package server

import (
	"context"

	agentproto "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/proto"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/server/imagescanservice"
)

func (s *AgentServer) ImageScanRun(
	ctx context.Context,
	req *agentproto.ImageScanRunReq,
) (*agentproto.ImageScanRunResp, error) {
	return imagescanservice.ImageScanRun(ctx, s.imageScanAgentUC, req) //nolint:wrapcheck
}
