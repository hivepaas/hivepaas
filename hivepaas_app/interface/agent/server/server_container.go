package server

import (
	agentproto "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/proto"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/server/containerservice"
)

func (s *AgentServer) ContainerExec(req agentproto.ContainerService_ContainerExecServer) error {
	return containerservice.ContainerExec(s.containerAgentUC, req) //nolint:wrapcheck
}

func (s *AgentServer) ContainerCopyFrom(
	req *agentproto.ContainerCopyFromReq,
	stream agentproto.ContainerService_ContainerCopyFromServer,
) error {
	return containerservice.ContainerCopyFrom(s.containerAgentUC, req, stream) //nolint:wrapcheck
}
