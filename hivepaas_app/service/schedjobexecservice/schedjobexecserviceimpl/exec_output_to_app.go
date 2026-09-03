package schedjobexecserviceimpl

import (
	"context"
	"io"

	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/safego"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/containerexecservice"
)

func (s *service) initOutputWriterToApp(
	ctx context.Context,
	data *execData,
) (writer io.WriteCloser, err error) {
	pipeToApp := data.SchedJob.CommandOutput.PipeToApp
	targetApp := data.RefObjects.RefApps[pipeToApp.TargetApp.ID]
	if targetApp == nil {
		return nil, hperrors.NewNotFound("Target app")
	}
	if targetApp.Status != base.AppStatusActive {
		return nil, hperrors.Wrap(hperrors.ErrAppInactive)
	}
	if targetApp.Project.Status != base.ProjectStatusActive {
		return nil, hperrors.Wrap(hperrors.ErrProjectInactive)
	}

	pr, pw := io.Pipe()
	data.uploadErrChan = make(chan error, 1)

	go func() {
		var finalErr error
		// NOTE: defers run LIFO. Catch a panic first so it becomes finalErr, then
		// close the pipe and always publish the result: skipping the send would
		// block the `<-data.uploadErrChan` read in exec_output.go forever.
		defer func() { data.uploadErrChan <- finalErr }()
		defer pr.Close()
		defer safego.RecoverTo(&finalErr)

		var calcErr error
		_, execErr := s.containerExecService.ContainerExec(ctx, &containerexecservice.ContainerExecReq{
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
				cmd, err := s.calcCommandHelper(ctx, pipeToApp.Command, data.Task.ID, data)
				if err != nil {
					calcErr = err
					return
				}
				opts.Cmd = cmd
				opts.WorkingDir = pipeToApp.Command.WorkingDir
			},
		})

		if calcErr != nil {
			finalErr = calcErr
		} else if execErr != nil {
			finalErr = execErr
		}
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
