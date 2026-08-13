package server

import (
	"context"
	"fmt"

	"github.com/hivepaas/hivepaas/hivepaas_app/infra/logging"
	agentproto "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/proto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/containeragentuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/imagebuildagentuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/nodecleanupagentuc"
)

type AgentServer struct {
	agentproto.UnimplementedAgentServiceServer
	agentproto.UnimplementedContainerServiceServer
	agentproto.UnimplementedNodeCleanupServiceServer
	agentproto.UnimplementedImageBuildServiceServer
	logger             logging.Logger
	containerAgentUC   *containeragentuc.UC
	imageBuildAgentUC  *imagebuildagentuc.UC
	nodeCleanupAgentUC *nodecleanupagentuc.UC
}

func NewAgentServer(
	logger logging.Logger,
	containerAgentUC *containeragentuc.UC,
	imageBuildAgentUC *imagebuildagentuc.UC,
	nodeCleanupAgentUC *nodecleanupagentuc.UC,
) *AgentServer {
	return &AgentServer{
		logger:             logger,
		containerAgentUC:   containerAgentUC,
		imageBuildAgentUC:  imageBuildAgentUC,
		nodeCleanupAgentUC: nodeCleanupAgentUC,
	}
}

func (s *AgentServer) Ping(
	_ context.Context,
	req *agentproto.PingReq,
) (*agentproto.PingResp, error) {
	s.logger.Infof("Received Ping request with message: %s", req.GetMessage())
	return &agentproto.PingResp{
		Message: fmt.Sprintf("Pong: %s", req.GetMessage()),
	}, nil
}
