package appdeploymentserviceimpl

import (
	"context"
	"fmt"
	"time"

	"github.com/moby/moby/client"
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/pkg/applog"
	"github.com/localpaas/localpaas/localpaas_app/pkg/shellutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/service/appdeploymentservice"
	"github.com/localpaas/localpaas/services/docker"
)

const (
	stepPreDeployCmd  = "pre-deploy-cmd-exec"
	stepPostDeployCmd = "post-deploy-cmd-exec"

	preDeploymentContainerFindRetryMax   = 1
	preDeploymentContainerFindRetryDelay = time.Second * 5

	postDeploymentContainerFindRetryMax   = 3
	postDeploymentContainerFindRetryDelay = time.Second * 10
)

func (s *service) deployStepExecCmd(
	ctx context.Context,
	data *appdeploymentservice.DeploymentData,
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
		_ = data.LogStore.Add(ctx, applog.NewWarnFrame(
			"No running container found, execution skipped", applog.TsNow))
		return nil
	}

	var cmdStr string
	if preDeployment {
		cmdStr = deployment.Settings.PreDeploymentCommand
	} else {
		cmdStr = deployment.Settings.PostDeploymentCommand
	}

	execInfo, logs, err := s.dockerManager.ContainerExecWait(ctx, contSum.ID, func(opts *client.ExecCreateOptions) {
		opts.AttachStdout = true
		opts.AttachStderr = true
		opts.Cmd = gofn.Must(shellutil.CmdSplit(cmdStr))
		opts.WorkingDir = deployment.Settings.WorkingDir
		opts.TTY = true
		opts.ConsoleSize = docker.DefaultConsoleSize
	})
	if err != nil {
		return apperrors.Wrap(err)
	}
	_ = data.LogStore.Add(ctx, logs...)

	if execInfo.ExitCode != 0 {
		_ = data.LogStore.Add(ctx, applog.NewErrFrame(fmt.Sprintf(
			"Command execution failed with exit code: %v", execInfo.ExitCode), applog.TsNow))
		return apperrors.Wrap(apperrors.ErrInfraActionFailed)
	}

	return nil
}
