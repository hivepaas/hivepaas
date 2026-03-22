package taskcronjobexec

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/api/types/container"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/applog"
	"github.com/localpaas/localpaas/localpaas_app/pkg/shellutil"
	"github.com/localpaas/localpaas/services/docker"
)

const (
	retryMax   = 3
	retryDelay = time.Second * 5
)

func (e *Executor) cronExecContainerCmd(
	ctx context.Context,
	db database.IDB,
	data *taskData,
) (err error) {
	command := data.CronJob.Command
	if command == nil || command.Command == "" { // can't continue if this happens
		data.NonRetryable = true
		_ = data.LogStore.Add(ctx, applog.NewErrFrame(
			"Execution command is empty, aborted", applog.TsNow))
		return apperrors.New(apperrors.ErrInternalServer).WithMsgLog("cron job command is empty")
	}

	contSum, _, err := e.dockerManager.ServiceContainerGetActive(ctx, data.App.ServiceID,
		retryMax, retryDelay)
	if err != nil {
		return apperrors.Wrap(err)
	}
	if contSum == nil {
		_ = data.LogStore.Add(ctx, applog.NewWarnFrame(
			"No running container found, execution skipped", applog.TsNow))
		return nil
	}

	envVars, err := e.cronJobService.BuildCommandEnv(ctx, db, data.App, data.CronJob)
	if err != nil {
		return apperrors.Wrap(err)
	}
	env := make([]string, 0, len(envVars))
	for _, v := range envVars {
		env = append(env, v.ToString("="))
	}

	var cmd []string
	if command.RunInShell != "" {
		cmd = []string{command.RunInShell, "-c", shellutil.ArgQuote(command.Command)}
	} else {
		cmd, err = shellutil.CmdSplit(command.Command)
		if err != nil {
			return apperrors.Wrap(err)
		}
	}

	execOptions := &container.ExecOptions{
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          cmd,
		WorkingDir:   command.WorkingDir,
		Env:          env,
		Tty:          true,
		ConsoleSize:  &docker.DefaultConsoleSize,
	}

	execInfo, logs, err := e.dockerManager.ContainerExecWait(ctx, contSum.ID, execOptions)
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
