package containerexecserviceimpl

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/moby/moby/client"
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/executil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/funcutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/redact"
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

	envVars, refSecrets, err := s.schedJobService.BuildCommandEnv(ctx, db, data.App, schedJob)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	env := make([]string, 0, len(envVars))
	for _, v := range envVars {
		env = append(env, v.ToString("="))
	}

	if len(refSecrets) > 0 && data.LogStore != nil {
		secrets := make([]string, 0, len(refSecrets))
		for _, secret := range refSecrets {
			secrets = append(secrets, secret.Value.MustGetPlain())
		}
		data.LogStore.SetRedactor(redact.New(secrets))
	}

	var cmd []string
	if command.RunInShell != "" {
		cmd = []string{command.RunInShell, "-c", executil.ArgQuote(command.Command)}
	} else {
		cmd, err = executil.CmdSplit(command.Command)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
	}

	createResp, attachResp, _, err := s.dockerManager.ContainerExec(ctx, contSum.ID, func(opts *client.ExecCreateOptions) {
		opts.AttachStdout = true
		opts.AttachStderr = true
		opts.Cmd = cmd
		opts.WorkingDir = command.WorkingDir
		opts.Env = env
		opts.TTY = command.TTY
		opts.ConsoleSize.Width = gofn.Coalesce(command.ConsoleSize.Width, docker.DefaultConsoleSize.Width)
		opts.ConsoleSize.Height = gofn.Coalesce(command.ConsoleSize.Height, docker.DefaultConsoleSize.Height)
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	defer attachResp.Close()

	logChan, _ := docker.StartScanningLog(ctx, io.NopCloser(attachResp.Reader), docker.WithParseLogHeader(false))

	for msgs := range logChan {
		_ = data.LogStore.AddRedacted(ctx, msgs...)
	}

	execInfo, err := s.dockerManager.ContainerExecInspect(ctx, createResp.ID)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	if execInfo.ExitCode != 0 {
		_ = data.LogStore.AddRedacted(ctx, tasklog.NewErrFrame(fmt.Sprintf(
			"Command execution failed with exit code: %v", execInfo.ExitCode), tasklog.TsNow))
		return nil, apperrors.Wrap(apperrors.ErrInfraActionFailed)
	}

	return resp, nil
}
