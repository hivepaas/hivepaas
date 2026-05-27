package containerexecserviceimpl

import (
	"context"
	"fmt"
	"time"

	"github.com/moby/moby/client"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/funcutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/shellutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/tasklog"
	"github.com/localpaas/localpaas/localpaas_app/service/containerexecservice"
	"github.com/localpaas/localpaas/services/docker"
)

const (
	retryMax   = 3
	retryDelay = time.Second * 5
)

func (s *service) ContainerExec(
	ctx context.Context,
	db database.Tx,
	data *containerexecservice.ContainerExecReq,
) (resp *containerexecservice.ContainerExecResp, err error) {
	defer funcutil.EnsureNoPanic(&err)

	resp = &containerexecservice.ContainerExecResp{}

	schedJob := data.SchedJob.MustAsSchedJob()
	command := schedJob.Command
	if command == nil || command.Command == "" { // can't continue if this happens
		data.TaskNonRetryable = true
		_ = data.LogStore.Add(ctx, tasklog.NewErrFrame(
			"Execution command is empty, aborted", tasklog.TsNow))
		return nil, apperrors.New(apperrors.ErrInternalServer).WithMsgLog("schedule job command is empty")
	}

	contSum, _, err := s.dockerManager.ServiceContainerGetActive(ctx, data.App.ServiceID,
		retryMax, retryDelay)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	if contSum == nil {
		_ = data.LogStore.Add(ctx, tasklog.NewWarnFrame(
			"No running container found, execution skipped", tasklog.TsNow))
		return resp, nil
	}

	envVars, err := s.schedJobService.BuildCommandEnv(ctx, db, data.App, schedJob)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	env := make([]string, 0, len(envVars))
	for _, v := range envVars {
		env = append(env, v.ToString("="))
	}

	var cmd []string
	if command.RunInShell != "" {
		cmd = []string{command.RunInShell, "-c", shellutil.ArgQuote(command.Command)}
	} else {
		cmd, err = shellutil.CmdSplit(command.Command)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
	}

	execInfo, logs, err := s.dockerManager.ContainerExecWait(ctx, contSum.ID, func(opts *client.ExecCreateOptions) {
		opts.AttachStdout = true
		opts.AttachStderr = true
		opts.Cmd = cmd
		opts.WorkingDir = command.WorkingDir
		opts.Env = env
		opts.TTY = true
		opts.ConsoleSize = docker.DefaultConsoleSize
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	_ = data.LogStore.Add(ctx, logs...)

	if execInfo.ExitCode != 0 {
		_ = data.LogStore.Add(ctx, tasklog.NewErrFrame(fmt.Sprintf(
			"Command execution failed with exit code: %v", execInfo.ExitCode), tasklog.TsNow))
		return nil, apperrors.Wrap(apperrors.ErrInfraActionFailed)
	}

	return resp, nil
}
