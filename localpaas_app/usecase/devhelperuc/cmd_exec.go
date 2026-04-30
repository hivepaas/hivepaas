package devhelperuc

import (
	"context"
	"os/exec"
	"strings"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/pkg/reflectutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/shellutil"
	"github.com/localpaas/localpaas/localpaas_app/usecase/devhelperuc/devhelperdto"
)

func (uc *UC) ExecuteCmd(
	ctx context.Context,
	auth *basedto.Auth,
	req *devhelperdto.ExecuteCmdReq,
) (_ *devhelperdto.ExecuteCmdResp, err error) {
	cmdArray := req.CmdArray
	if req.Cmd != "" {
		cmdArray, err = shellutil.CmdSplit(req.Cmd)
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
