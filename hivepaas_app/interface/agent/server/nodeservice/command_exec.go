package nodeservice

import (
	"bytes"
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	agentproto "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/proto"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/nodeexecservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/nodeagentuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/nodeagentuc/nodeagentdto"
)

func ExecuteCommand(
	ctx context.Context,
	uc *nodeagentuc.UC,
	req *agentproto.ExecCommandReq,
) (*agentproto.ExecCommandResp, error) {
	if req == nil {
		return &agentproto.ExecCommandResp{}, nil
	}

	var stdoutBuf, stderrBuf bytes.Buffer

	dtoReq := &nodeagentdto.ExecCommandReq{
		CommandExecOpts: &nodeexecservice.CommandExecOpts{
			Command:    req.GetCommand(),
			Env:        req.GetEnv(),
			WorkingDir: req.GetWorkingDir(),
			Stdout:     &stdoutBuf,
			Stderr:     &stderrBuf,
		},
	}

	resp, err := uc.ExecuteCommand(ctx, dtoReq)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if resp == nil {
		return &agentproto.ExecCommandResp{}, nil
	}

	return &agentproto.ExecCommandResp{
		Stdout:   stdoutBuf.Bytes(),
		Stderr:   stderrBuf.Bytes(),
		ExitCode: resp.ExitCode,
	}, nil
}
