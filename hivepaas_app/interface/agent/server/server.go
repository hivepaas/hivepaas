package server

import (
	"context"
	"fmt"

	"github.com/hivepaas/hivepaas/hivepaas_app/infra/logging"
	agentproto "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/proto"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/server/containerservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/containeragentuc"
)

// AgentServer implements agentproto.AgentServiceServer and agentproto.ContainerServiceServer
type AgentServer struct {
	agentproto.UnimplementedAgentServiceServer
	agentproto.UnimplementedContainerServiceServer
	logger           logging.Logger
	containerAgentUC *containeragentuc.UC
}

func NewAgentServer(
	logger logging.Logger,
	containerAgentUC *containeragentuc.UC,
) *AgentServer {
	return &AgentServer{
		logger:           logger,
		containerAgentUC: containerAgentUC,
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

/// CONTAINER SERVICE

func (s *AgentServer) ContainerExec(req agentproto.ContainerService_ContainerExecServer) error {
	return containerservice.ContainerExec(s.containerAgentUC, req) //nolint:wrapcheck
}
