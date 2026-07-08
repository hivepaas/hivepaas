package schedjobexecserviceimpl

import (
	"context"
	"io"

	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/funcutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/containerexecservice"
)

func (s *service) initOutputWriterToApp(
	ctx context.Context,
	data *execData,
) (writer io.WriteCloser, err error) {
	pipeToApp := data.SchedJob.CommandOutput.PipeToApp
	targetApp := data.RefObjects.RefApps[pipeToApp.TargetApp.ID]
	if targetApp == nil {
		return nil, apperrors.NewNotFound("Target app")
	}
	if targetApp.Status != base.AppStatusActive {
		return nil, apperrors.New(apperrors.ErrAppInactive)
	}

	targetProject := data.Project
	if targetApp.ProjectID != targetProject.ID {
		targetProject = targetApp.Project
	}
	if targetProject.Status != base.ProjectStatusActive {
		return nil, apperrors.New(apperrors.ErrProjectInactive)
	}

	pr, pw := io.Pipe()
	data.uploadErrChan = make(chan error, 1)

	go func() {
		defer funcutil.EnsureNoPanic(nil)
		defer pr.Close()

		var calcErr error
		_, execErr := s.containerExecService.ContainerExec(ctx, &containerexecservice.ContainerExecReq{
			Project:                targetProject,
			App:                    targetApp,
			TaskMinRunningDuration: data.TaskMinRunningDuration,
			TaskFindRetryMax:       data.TaskFindRetryMax,
			TaskFindRetryDelay:     data.TaskFindRetryDelay,
			LogStore:               data.LogStore,
			StdinReader:            pr,
			ExecOptions: func(opts *client.ExecCreateOptions) {
				opts.AttachStdin = true
				opts.AttachStdout = true
				opts.AttachStderr = true
				cmd, err := s.calcCommandHelper(ctx, pipeToApp.Command, data.Task.ID, data.LogStore)
				if err != nil {
					calcErr = err
					return
				}
				opts.Cmd = cmd
				opts.WorkingDir = pipeToApp.Command.WorkingDir
			},
		})

		var finalErr error
		if calcErr != nil {
			finalErr = calcErr
		} else if execErr != nil {
			finalErr = execErr
		}
		data.uploadErrChan <- finalErr
	}()

	baseWriter := &writeCloserWrapper{
		Writer:    pw,
		closeFunc: func() error { return pw.Close() },
	}
	data.closeStack = func() error {
		return baseWriter.Close()
	}
	return baseWriter, nil
}
