package server

import (
	agentproto "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/proto"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/server/imagebuildservice"
)

func (s *AgentServer) ImageBuild(
	req *agentproto.ImageBuildReq,
	stream agentproto.ImageBuildService_ImageBuildServer,
) error {
	return imagebuildservice.ImageBuild(s.imageBuildAgentUC, req, stream) //nolint:wrapcheck
}
