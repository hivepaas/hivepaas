package server

import (
	"context"
	"fmt"

	agentproto "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/proto"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/containeragentuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/dockerapiagentuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/imagebuildagentuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/nodeagentuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/nodecleanupagentuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/volumeagentuc"
)

type AgentServer struct {
	agentproto.UnimplementedAgentServiceServer
	agentproto.UnimplementedContainerServiceServer
	agentproto.UnimplementedDockerAPIServiceServer
	agentproto.UnimplementedImageBuildServiceServer
	agentproto.UnimplementedNodeCleanupServiceServer
	agentproto.UnimplementedNodeServiceServer
	agentproto.UnimplementedVolumeServiceServer
	logger             logging.Logger
	containerAgentUC   *containeragentuc.UC
	dockerAPIAgentUC   *dockerapiagentuc.UC
	imageBuildAgentUC  *imagebuildagentuc.UC
	nodeAgentUC        *nodeagentuc.UC
	nodeCleanupAgentUC *nodecleanupagentuc.UC
	volumeAgentUC      *volumeagentuc.UC
}

func NewAgentServer(
	logger logging.Logger,
	containerAgentUC *containeragentuc.UC,
	dockerAPIAgentUC *dockerapiagentuc.UC,
	imageBuildAgentUC *imagebuildagentuc.UC,
	nodeAgentUC *nodeagentuc.UC,
	nodeCleanupAgentUC *nodecleanupagentuc.UC,
	volumeAgentUC *volumeagentuc.UC,

) *AgentServer {
	return &AgentServer{
		logger:             logger,
		containerAgentUC:   containerAgentUC,
		dockerAPIAgentUC:   dockerAPIAgentUC,
		imageBuildAgentUC:  imageBuildAgentUC,
		nodeAgentUC:        nodeAgentUC,
		nodeCleanupAgentUC: nodeCleanupAgentUC,
		volumeAgentUC:      volumeAgentUC,
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
