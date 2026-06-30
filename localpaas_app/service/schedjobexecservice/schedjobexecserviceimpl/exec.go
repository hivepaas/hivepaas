package schedjobexecserviceimpl

import (
	"context"
	"io"
	"time"

	"github.com/moby/moby/client"
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/funcutil"
	"github.com/localpaas/localpaas/localpaas_app/service/containerexecservice"
	"github.com/localpaas/localpaas/localpaas_app/service/schedjobexecservice"
	"github.com/localpaas/localpaas/services/docker"
)

type schedJobExecData struct {
	*schedjobexecservice.SchedJobExecReq

	SchedJob *entity.SchedJob
	File     *entity.File
	TimeNow  time.Time

	uploadFunc    func(_ context.Context, objectKey string, data io.Reader) error
	uploadErrChan chan error
	closeStack    func() error
}

func (s *service) SchedJobExec(
	ctx context.Context,
	db database.Tx,
	req *schedjobexecservice.SchedJobExecReq,
) (_ *schedjobexecservice.SchedJobExecResp, err error) {
	defer funcutil.EnsureNoPanic(&err)

	schedJob := req.SchedJobSetting.MustAsSchedJob()
	command := schedJob.Command
	data := &schedJobExecData{
		SchedJobExecReq: req,
		SchedJob:        schedJob,
		TimeNow:         time.Now(),
	}

	cmd, err := s.schedJobExecCalcCommand(ctx, data)
	if err != nil {
		return nil, apperrors.New(err)
	}

	env, err := s.schedJobExecCalcCommandEnv(ctx, db, data)
	if err != nil {
		return nil, apperrors.New(err)
	}

	stdoutWriter, err := s.schedJobExecInitWriter(ctx, data)
	if err != nil {
		return nil, apperrors.New(err)
	}

	defer s.schedJobExecCleanup(err, data)

	_, err = s.containerExecService.ContainerExec(ctx, &containerexecservice.ContainerExecReq{
		Project:                req.Project,
		App:                    req.App,
		TaskMinRunningDuration: req.TaskMinRunningDuration,
		TaskFindRetryMax:       req.TaskFindRetryMax,
		TaskFindRetryDelay:     req.TaskFindRetryDelay,
		LogStore:               req.LogStore,
		StdoutWriter:           stdoutWriter,
		ExecOptions: func(opts *client.ExecCreateOptions) {
			opts.AttachStdout = true
			opts.AttachStderr = true
			opts.Cmd = cmd
			opts.WorkingDir = command.WorkingDir
			opts.Env = env
			// NOTE: when redirect command stdout to a custom writer, we set TTY=false
			if stdoutWriter == nil {
				opts.TTY = command.TTY
				opts.ConsoleSize.Width = gofn.Coalesce(command.ConsoleSize.Width, docker.DefaultConsoleSize.Width)
				opts.ConsoleSize.Height = gofn.Coalesce(command.ConsoleSize.Height, docker.DefaultConsoleSize.Height)
			}
		},
	})

	err = s.schedJobExecFinalize(ctx, db, err, data)
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &schedjobexecservice.SchedJobExecResp{}, nil
}
