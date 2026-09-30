package server

import (
	"google.golang.org/grpc"

	agentproto "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/proto"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/server/imagebuildservice"
)

func (s *AgentServer) ImageBuildFromSource(
	stream grpc.BidiStreamingServer[agentproto.ImageBuildMsg, agentproto.ImageBuildResp],
) error {
	return imagebuildservice.ImageBuildFromSource(s.imageBuildAgentUC, stream) //nolint:wrapcheck
}
