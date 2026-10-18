package appdeploymentserviceimpl

import (
	"context"
	"fmt"
	"time"

	"github.com/moby/moby/client"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/executil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/containerexecservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

const (
	stepPreDeployCmd  = "pre-deploy-cmd-exec"
	stepPostDeployCmd = "post-deploy-cmd-exec"

	preDeploymentTaskFindRetryMax       = 3
	preDeploymentTaskFindRetryDelay     = time.Second * 5
	preDeploymentTaskMinRunningDuration = time.Second * 15

	postDeploymentTaskFindRetryMax       = 10
	postDeploymentTaskFindRetryDelay     = time.Second * 5
	postDeploymentTaskMinRunningDuration = time.Second * 20
)

func (s *service) deployStepExecCmd(
	ctx context.Context,
	data *appDeploymentData,
	preDeployment bool,
) (err error) {
	deployment := data.Deployment
	cmdStr, step := deployment.Settings.PostDeploymentCommand, stepPostDeployCmd
	if preDeployment {
		cmdStr, step = deployment.Settings.PreDeploymentCommand, stepPreDeployCmd
	}
	if cmdStr == "" {
		return nil
	}
	data.Step = step

	s.addStepStartLog(ctx, data, fmt.Sprintf("Start executing %s-deployment command...",
		gofn.If(preDeployment, "pre", "post")))
	defer s.addStepEndLog(ctx, data, timeutil.NowUTC(), err)

	var taskFindRetryMax int
	var taskFindRetryDelay time.Duration
	var taskMinRunningDuration time.Duration
	if preDeployment {
		taskFindRetryMax = preDeploymentTaskFindRetryMax
		taskFindRetryDelay = preDeploymentTaskFindRetryDelay
		taskMinRunningDuration = preDeploymentTaskMinRunningDuration
	} else {
		taskFindRetryMax = postDeploymentTaskFindRetryMax
		taskFindRetryDelay = postDeploymentTaskFindRetryDelay
		taskMinRunningDuration = postDeploymentTaskMinRunningDuration
	}

	_, err = s.containerExecService.ContainerExec(ctx, &containerexecservice.ContainerExecReq{
		App:                    data.App,
		TaskMinRunningDuration: taskMinRunningDuration,
		TaskFindRetryMax:       taskFindRetryMax,
		TaskFindRetryDelay:     taskFindRetryDelay,
		LogStore:               data.LogStore,
		ExecOptions: func(opts *client.ExecCreateOptions) {
			opts.AttachStdout = true
			opts.AttachStderr = true
			opts.Cmd = gofn.Must(executil.CmdSplit(cmdStr))
			opts.WorkingDir = deployment.Settings.WorkingDir
			opts.TTY = true
			opts.ConsoleSize = docker.DefaultConsoleSize
		},
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
