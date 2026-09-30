package server

import (
	"context"
	"fmt"

	agentproto "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/proto"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/containeragentuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/dockerapiagentuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/fileagentuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/imagebuildagentuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/nodeagentuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/nodecleanupagentuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/reposerveragentuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/volumeagentuc"
)

type AgentServer struct {
	agentproto.UnimplementedAgentServiceServer
	agentproto.UnimplementedContainerServiceServer
	agentproto.UnimplementedDockerAPIServiceServer
	agentproto.UnimplementedFileServiceServer
	agentproto.UnimplementedImageBuildServiceServer
	agentproto.UnimplementedNodeCleanupServiceServer
	agentproto.UnimplementedNodeServiceServer
	agentproto.UnimplementedRepoServerServiceServer
	agentproto.UnimplementedVolumeServiceServer
	logger             logging.Logger
	containerAgentUC   *containeragentuc.UC
	dockerAPIAgentUC   *dockerapiagentuc.UC
	fileAgentUC        *fileagentuc.UC
	imageBuildAgentUC  *imagebuildagentuc.UC
	nodeAgentUC        *nodeagentuc.UC
	nodeCleanupAgentUC *nodecleanupagentuc.UC
	repoServerAgentUC  *reposerveragentuc.UC
	volumeAgentUC      *volumeagentuc.UC
}

func NewAgentServer(
	logger logging.Logger,
	containerAgentUC *containeragentuc.UC,
	dockerAPIAgentUC *dockerapiagentuc.UC,
	fileAgentUC *fileagentuc.UC,
	imageBuildAgentUC *imagebuildagentuc.UC,
	nodeAgentUC *nodeagentuc.UC,
	nodeCleanupAgentUC *nodecleanupagentuc.UC,
	repoServerAgentUC *reposerveragentuc.UC,
	volumeAgentUC *volumeagentuc.UC,

) *AgentServer {
	return &AgentServer{
		logger:             logger,
		containerAgentUC:   containerAgentUC,
		dockerAPIAgentUC:   dockerAPIAgentUC,
		fileAgentUC:        fileAgentUC,
		imageBuildAgentUC:  imageBuildAgentUC,
		nodeAgentUC:        nodeAgentUC,
		nodeCleanupAgentUC: nodeCleanupAgentUC,
		repoServerAgentUC:  repoServerAgentUC,
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
