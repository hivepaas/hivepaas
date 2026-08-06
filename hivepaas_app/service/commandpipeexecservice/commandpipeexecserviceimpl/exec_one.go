package commandpipeexecserviceimpl

import (
	"context"
	"fmt"
	"io"

	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/funcutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/containerexecservice"
)

func (s *service) commandPipeExec(
	ctx context.Context,
	db database.IDB,
	pipeSetting *entity.Setting,
	data *execData,
) (err error) {
	cmdPipe, err := pipeSetting.AsCommandPipe()
	if err != nil {
		return apperrors.Wrap(err)
	}

	var srcCmdSetting, destCmdSetting *entity.Setting
	if cmdPipe.SourceCommand.ID != "" {
		srcCmdSetting = data.RefObjects.RefSettings[cmdPipe.SourceCommand.ID]
		if srcCmdSetting == nil {
			return apperrors.NewNotFound("Source command template")
		}
	}
	if cmdPipe.TargetCommand.ID != "" {
		destCmdSetting = data.RefObjects.RefSettings[cmdPipe.TargetCommand.ID]
		if destCmdSetting == nil {
			return apperrors.NewNotFound("Target command template")
		}
	}

	if srcCmdSetting == nil && destCmdSetting == nil {
		return apperrors.NewArgumentInvalid("source or target command")
	}
	if destCmdSetting == nil {
		return s.singleCommandExec(ctx, db, data.SrcApp, srcCmdSetting, data)
	}
	if srcCmdSetting == nil {
		return s.singleCommandExec(ctx, db, data.DestApp, destCmdSetting, data)
	}
	return s.doCommandPipeExec(ctx, db, pipeSetting, srcCmdSetting, destCmdSetting, data)
}

func (s *service) singleCommandExec(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	cmdSetting *entity.Setting,
	data *execData,
) (err error) {
	cmdTemplate := cmdSetting.MustAsCommandTemplate()

	cmd, err := s.calcCommand(ctx, cmdTemplate, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	env, err := s.calcCommandEnv(ctx, db, app, cmdTemplate, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	_ = data.LogStore.Add(ctx, tasklog.NewOutFrame(
		fmt.Sprintf("Executing single command '%s' on app [%s]...", cmdSetting.Name, app.Name),
		tasklog.TsNow,
	))

	_, err = s.containerExecService.ContainerExec(ctx, &containerexecservice.ContainerExecReq{
		App:                    app,
		TaskMinRunningDuration: data.TaskMinRunningDuration,
		TaskFindRetryMax:       data.TaskFindRetryMax,
		TaskFindRetryDelay:     data.TaskFindRetryDelay,
		LogStore:               data.LogStore,
		ExecOptions: func(opts *client.ExecCreateOptions) {
			opts.AttachStdout = true
			opts.AttachStderr = true
			opts.Cmd = cmd
			opts.WorkingDir = cmdTemplate.WorkingDir
			opts.Env = env
			opts.TTY = false
		},
	})
	if err != nil {
		return apperrors.Wrap(err)
	}

	_ = data.LogStore.Add(ctx, tasklog.NewOutFrame(
		fmt.Sprintf("Single command '%s' completed successfully", cmdSetting.Name),
		tasklog.TsNow,
	))
	return nil
}

func (s *service) doCommandPipeExec(
	ctx context.Context,
	db database.IDB,
	pipeSetting *entity.Setting,
	srcCmdSetting, destCmdSetting *entity.Setting,
	data *execData,
) (err error) {
	srcCmdTemplate := srcCmdSetting.MustAsCommandTemplate()
	destCmdTemplate := destCmdSetting.MustAsCommandTemplate()

	srcCmd, err := s.calcCommand(ctx, srcCmdTemplate, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	destCmd, err := s.calcCommand(ctx, destCmdTemplate, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	srcEnv, err := s.calcCommandEnv(ctx, db, data.SrcApp, srcCmdTemplate, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	destEnv, err := s.calcCommandEnv(ctx, db, data.DestApp, destCmdTemplate, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	_ = data.LogStore.Add(ctx, tasklog.NewOutFrame(
		fmt.Sprintf("Executing command pipe '%s': [%s] '%s' -> [%s] '%s'...",
			pipeSetting.Name, data.SrcApp.Name, srcCmdSetting.Name, data.DestApp.Name, destCmdSetting.Name),
		tasklog.TsNow,
	))

	pr, pw := io.Pipe()
	errChan := make(chan error, 2) //nolint:mnd

	// 1. Source command execution (stdout written to pw)
	go func() {
		defer funcutil.EnsureNoPanic(nil)
		_, execErr := s.containerExecService.ContainerExec(ctx, &containerexecservice.ContainerExecReq{
			App:                    data.SrcApp,
			TaskMinRunningDuration: data.TaskMinRunningDuration,
			TaskFindRetryMax:       data.TaskFindRetryMax,
			TaskFindRetryDelay:     data.TaskFindRetryDelay,
			LogStore:               data.LogStore,
			StdoutWriter:           pw,
			ExecOptions: func(opts *client.ExecCreateOptions) {
				opts.AttachStdout = true
				opts.AttachStderr = true
				opts.Cmd = srcCmd
				opts.WorkingDir = srcCmdTemplate.WorkingDir
				opts.Env = srcEnv
				opts.TTY = false
			},
		})
		_ = pw.CloseWithError(execErr)
		errChan <- execErr
	}()

	// 2. Target command execution (stdin read from pr)
	go func() {
		defer funcutil.EnsureNoPanic(nil)
		_, execErr := s.containerExecService.ContainerExec(ctx, &containerexecservice.ContainerExecReq{
			App:                    data.DestApp,
			TaskMinRunningDuration: data.TaskMinRunningDuration,
			TaskFindRetryMax:       data.TaskFindRetryMax,
			TaskFindRetryDelay:     data.TaskFindRetryDelay,
			LogStore:               data.LogStore,
			StdinReader:            pr,
			ExecOptions: func(opts *client.ExecCreateOptions) {
				opts.AttachStdin = true
				opts.AttachStdout = true
				opts.AttachStderr = true
				opts.Cmd = destCmd
				opts.WorkingDir = destCmdTemplate.WorkingDir
				opts.Env = destEnv
				opts.TTY = false
			},
		})
		_ = pr.CloseWithError(execErr)
		errChan <- execErr
	}()

	err1 := <-errChan
	err2 := <-errChan

	if err1 != nil {
		return apperrors.Wrap(err1)
	}
	if err2 != nil {
		return apperrors.Wrap(err2)
	}

	_ = data.LogStore.Add(ctx, tasklog.NewOutFrame(
		fmt.Sprintf("Command pipe '%s' completed successfully", pipeSetting.Name),
		tasklog.TsNow,
	))
	return nil
}
