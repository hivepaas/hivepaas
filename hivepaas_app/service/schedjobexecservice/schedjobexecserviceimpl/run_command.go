package schedjobexecserviceimpl

import (
	"context"
	"time"

	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/safego"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/containerexecservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobexecservice"
)

func (s *service) RunCommand(
	ctx context.Context,
	db database.Tx,
	req *schedjobexecservice.RunCommandReq,
) (_ *schedjobexecservice.RunCommandResp, err error) {
	defer safego.RecoverTo(&err)

	data := &execData{
		SchedJobExecReq: &schedjobexecservice.SchedJobExecReq{TaskExecData: req.TaskExecData, DestApp: req.App},
		Command:         req.Command,
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

	execResp, err := s.containerExecService.ContainerExec(ctx, &containerexecservice.ContainerExecReq{
		App:         req.App,
		LogStore:    req.LogStore,
		StdinReader: req.Stdin,
		ExecOptions: func(opts *client.ExecCreateOptions) {
			opts.AttachStdout = true
			opts.AttachStderr = true
			opts.Cmd = cmd
			opts.WorkingDir = req.Command.WorkingDir
			opts.Env = env
			// What it reads is data: a TTY would not pass it through as is.
			opts.TTY = false
		},
	})
	resp := &schedjobexecservice.RunCommandResp{}
	if execResp != nil && execResp.ExecStarted {
		exitCode := execResp.ExitCode
		resp.ExitCode = &exitCode
	}
	if err != nil {
		return resp, hperrors.Wrap(err)
	}
	return resp, nil
}
