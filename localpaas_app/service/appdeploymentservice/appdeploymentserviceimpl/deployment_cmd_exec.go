package appdeploymentserviceimpl

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/moby/moby/client"
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/pkg/executil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/tasklog"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/services/docker"
)

const (
	stepPreDeployCmd  = "pre-deploy-cmd-exec"
	stepPostDeployCmd = "post-deploy-cmd-exec"

	preDeploymentContainerFindRetryMax   = 3
	preDeploymentContainerFindRetryDelay = time.Second * 5

	postDeploymentContainerFindRetryMax   = 10
	postDeploymentContainerFindRetryDelay = time.Second * 5
)

func (s *service) deployStepExecCmd(
	ctx context.Context,
	data *appDeploymentData,
	preDeployment bool,
) (err error) {
	deployment := data.Deployment
	if preDeployment && deployment.Settings.PreDeploymentCommand == "" {
		return nil
	} else {
		data.Step = stepPreDeployCmd
	}
	if !preDeployment && deployment.Settings.PostDeploymentCommand == "" {
		return nil
	} else {
		data.Step = stepPostDeployCmd
	}

	s.addStepStartLog(ctx, data, fmt.Sprintf("Start executing %s-deployment command...",
		gofn.If(preDeployment, "pre", "post")))
	defer s.addStepEndLog(ctx, data, timeutil.NowUTC(), err)

	var maxRetry int
	var retryDelay time.Duration
	if preDeployment {
		maxRetry = preDeploymentContainerFindRetryMax
		retryDelay = preDeploymentContainerFindRetryDelay
	} else {
		maxRetry = postDeploymentContainerFindRetryMax
		retryDelay = postDeploymentContainerFindRetryDelay
	}

	contSum, _, err := s.dockerManager.ServiceContainerGetActive(ctx, data.App.ServiceID,
		maxRetry, retryDelay)
	if err != nil {
		return apperrors.Wrap(err)
	}
	if contSum == nil {
		_ = data.LogStore.Add(ctx, tasklog.NewWarnFrame(
			"No running container found, execution skipped", tasklog.TsNow))
		return nil
	}

	var cmdStr string
	if preDeployment {
		cmdStr = deployment.Settings.PreDeploymentCommand
	} else {
		cmdStr = deployment.Settings.PostDeploymentCommand
	}

	createResp, attachResp, _, err := s.dockerManager.ContainerExec(ctx, contSum.ID, func(opts *client.ExecCreateOptions) {
		opts.AttachStdout = true
		opts.AttachStderr = true
		opts.Cmd = gofn.Must(executil.CmdSplit(cmdStr))
		opts.WorkingDir = deployment.Settings.WorkingDir
		opts.TTY = true
		opts.ConsoleSize = docker.DefaultConsoleSize
	})
	if err != nil {
		return apperrors.Wrap(err)
	}
	defer attachResp.Close()

	logChan, _ := docker.StartScanningLog(ctx, io.NopCloser(attachResp.Reader), docker.WithParseLogHeader(false))

	for msgs := range logChan {
		_ = data.LogStore.AddRedacted(ctx, msgs...)
	}

	execInfo, err := s.dockerManager.ContainerExecInspect(ctx, createResp.ID)
	if err != nil {
		return apperrors.Wrap(err)
	}

	if execInfo.ExitCode != 0 {
		_ = data.LogStore.Add(ctx, tasklog.NewErrFrame(fmt.Sprintf(
			"Command execution failed with exit code: %v", execInfo.ExitCode), tasklog.TsNow))
		return apperrors.Wrap(apperrors.ErrInfraActionFailed)
	}

	return nil
}
