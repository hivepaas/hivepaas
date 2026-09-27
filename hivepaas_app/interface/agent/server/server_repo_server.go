package server

import (
	agentproto "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/proto"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/server/reposerverservice"
)

func (s *AgentServer) RepoServer(stream agentproto.RepoServerService_RepoServerServer) error {
	return reposerverservice.RepoServer(s.repoServerAgentUC, stream) //nolint:wrapcheck
}
