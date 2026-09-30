package schedjobexecserviceimpl

import (
	"context"
	"io"
	"time"

	"github.com/moby/moby/client"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/safego"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/containerexecservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobexecservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

type execData struct {
	*schedjobexecservice.SchedJobExecReq

	SchedJob *entity.SchedJob
	// Command is what runs: the job's own, or the one the request gave.
	Command *entity.CommandTemplate
	File    *entity.File
	TimeNow time.Time
	// db reads what the refs do not hold, such as a storage's key auth.
	db database.IDB

	uploadFunc    func(_ context.Context, objectKey string, data io.Reader) error
	uploadErrChan chan error
	closeStack    func() error
}

func (s *service) SchedJobExec(
	ctx context.Context,
	db database.Tx,
	req *schedjobexecservice.SchedJobExecReq,
) (_ *schedjobexecservice.SchedJobExecResp, err error) {
	defer safego.RecoverTo(&err)

	schedJob := req.SchedJobSetting.MustAsSchedJob()
	command := schedJob.Command
	if req.Command != nil {
		command = req.Command
	}
	data := &execData{
		SchedJobExecReq: req,
		SchedJob:        schedJob,
		Command:         command,
		TimeNow:         time.Now(),
		db:              db,
	}

	cmd, err := s.calcCommand(ctx, data)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	env, err := s.calcCommandEnv(ctx, db, data)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	env = append(env, sequenceEnv(req.Sequence)...)
	env = append(env, triggerEnv(req.Task)...)

	var stdoutWriter io.Writer
	if req.StdoutWriter != nil {
		stdoutWriter = req.StdoutWriter
	} else {
		outputWriter, err := s.initOutputWriter(ctx, data)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		if outputWriter != nil {
			stdoutWriter = outputWriter
		}
	}

	// Deferred through a closure, so it sees the error the run ends with rather
	// than the one it had when deferred.
	defer func() { s.cleanup(ctx, err, data) }()

	execResp, err := s.containerExecService.ContainerExec(ctx, &containerexecservice.ContainerExecReq{
		App:                    req.DestApp,
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
				if command.TTY {
					opts.ConsoleSize.Width = gofn.Coalesce(command.ConsoleSize.Width, docker.DefaultConsoleSize.Width)
					opts.ConsoleSize.Height = gofn.Coalesce(command.ConsoleSize.Height, docker.DefaultConsoleSize.Height)
				}
			}
		},
	})

	resp := &schedjobexecservice.SchedJobExecResp{}
	if execResp != nil && execResp.ExecStarted {
		exitCode := execResp.ExitCode
		resp.ExitCode = &exitCode
	}
	if req.Sequence != nil {
		// Read even after a failure: a step may say why in its outputs.
		resp.Outputs = s.readOutputs(ctx, req, execResp)
	}

	err = s.finalize(ctx, db, err, data)
	if err != nil {
		return resp, hperrors.Wrap(err)
	}
	return resp, nil
}
