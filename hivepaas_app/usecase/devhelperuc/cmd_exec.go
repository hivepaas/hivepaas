package devhelperuc

import (
	"context"
	"os/exec"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/executil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/reflectutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/devhelperuc/devhelperdto"
)

func (uc *UC) ExecuteCmd(
	ctx context.Context,
	auth *basedto.Auth,
	req *devhelperdto.ExecuteCmdReq,
) (_ *devhelperdto.ExecuteCmdResp, err error) {
	cmdArray := req.CmdArray
	if req.Cmd != "" {
		cmdArray, err = executil.CmdSplit(req.Cmd)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
	}

	cmd := exec.Command(cmdArray[0], cmdArray[1:]...) //nolint:gosec
	cmd.Dir = req.Dir
	cmd.Env = req.Env

	resp := &devhelperdto.ExecuteCmdDataResp{}

	res, err := cmd.CombinedOutput()
	if err != nil {
		resp.Error = err.Error()
	}
	if cmd.ProcessState != nil {
		resp.ExitCode = cmd.ProcessState.ExitCode()
	}
	resp.Output = append(resp.Output, strings.Split(reflectutil.UnsafeBytesToStr(res), "\n")...)

	return &devhelperdto.ExecuteCmdResp{
		Data: resp,
	}, nil
}
