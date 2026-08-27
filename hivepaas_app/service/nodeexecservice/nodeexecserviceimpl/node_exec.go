package nodeexecserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/client/nodeservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/nodeexecservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/nodeagentuc/nodeagentdto"
)

func (s *service) ExecCommand(
	ctx context.Context,
	req *nodeexecservice.CommandExecReq,
) (resp *nodeexecservice.CommandExecResp, err error) {
	var agentAddr string
	if req.NodeID != "" {
		agentAddr, err = s.agentService.GetAgentAddrForNode(ctx, req.NodeID)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
	}
	if agentAddr == "" && req.NodeLabel != "" {
		agentAddr, err = s.agentService.GetAgentAddrForNodeLabel(ctx, req.NodeLabel)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
	}

	agentClient, err := nodeservice.NewNodeServiceClient(agentAddr)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	defer agentClient.Close()

	agentReq := &nodeagentdto.ExecCommandReq{
		CommandExecOpts: req.CommandExecOpts,
	}

	execResp, err := agentClient.ExecuteCommand(ctx, agentReq)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &nodeexecservice.CommandExecResp{
		ExitCode: execResp.ExitCode,
	}, nil
}
