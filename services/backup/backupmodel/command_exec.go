package backupmodel

import (
	"context"
	"io"
	"os/exec"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/envutil"
)

type CommandExecReq struct {
	Command    []string
	Env        []string
	WorkingDir string
	Stdin      io.Reader
	Stdout     io.Writer
	Stderr     io.Writer // if nil, error data will go through Stdout, use io.Discard to discard

	NodeID    string // if empty, use NodeLabel
	NodeLabel string // if empty, execute command in the current node
}

type CommandExecResp struct {
	ExitCode int32
}

type CommandExecutor func(ctx context.Context, req *CommandExecReq) (*CommandExecResp, error)

func DefaultCommandExecutor(
	ctx context.Context,
	req *CommandExecReq,
) (*CommandExecResp, error) {
	if len(req.Command) == 0 {
		return nil, hperrors.Wrap(ErrCommandRequired)
	}

	cmd := exec.CommandContext(ctx, req.Command[0], req.Command[1:]...) //nolint:gosec
	if len(req.Env) > 0 {
		cmd.Env = append(envutil.SafeEnviron(), req.Env...)
	}
	if req.WorkingDir != "" {
		cmd.Dir = req.WorkingDir
	}
	if req.Stdin != nil {
		cmd.Stdin = req.Stdin
	}
	if req.Stdout != nil {
		cmd.Stdout = req.Stdout
	}
	if req.Stderr != nil {
		cmd.Stderr = req.Stderr
	}

	err := cmd.Run()
	var exitCode int32
	if cmd.ProcessState != nil {
		exitCode = int32(cmd.ProcessState.ExitCode()) //nolint:gosec
	}
	if err != nil {
		return &CommandExecResp{ExitCode: exitCode}, hperrors.Wrap(err)
	}

	return &CommandExecResp{ExitCode: exitCode}, nil
}
