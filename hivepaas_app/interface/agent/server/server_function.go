package server

import (
	"context"

	agentproto "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/proto"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/server/functionservice"
)

func (s *AgentServer) FunctionTestRun(
	ctx context.Context,
	req *agentproto.FunctionTestRunReq,
) (*agentproto.FunctionTestRunResp, error) {
	return functionservice.FunctionTestRun(ctx, s.functionAgentUC, req) //nolint:wrapcheck
}
