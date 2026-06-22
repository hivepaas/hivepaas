package containerexecserviceimpl

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/moby/moby/client"
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/executil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/funcutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/reflectutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/tasklog"
	"github.com/localpaas/localpaas/localpaas_app/service/containerexecservice"
	"github.com/localpaas/localpaas/services/docker"
)

func (s *service) SchedJobExec(
	ctx context.Context,
	db database.Tx,
	req *containerexecservice.SchedJobExecReq,
) (_ *containerexecservice.SchedJobExecResp, err error) {
	defer funcutil.EnsureNoPanic(&err)

	schedJob := req.SchedJobSetting.MustAsSchedJob()
	command := schedJob.Command
	if command == nil || (command.Command == "" && command.Script == "") { // can't continue if this happens
		req.TaskNonRetryable = true
		_ = req.LogStore.Add(ctx, tasklog.NewErrFrame(
			"Execution command/script is empty, aborted", tasklog.TsNow))
		return nil, apperrors.New(apperrors.ErrInternalServer).WithMsgLog("schedule job command/script is empty")
	}

	envVars, refSecrets, err := s.schedJobService.BuildCommandEnv(ctx, db, req.App, schedJob)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	env := make([]string, 0, len(envVars))
	for _, v := range envVars {
		env = append(env, v.ToString("="))
	}

	if len(refSecrets) > 0 && req.LogStore != nil {
		secrets := make([]string, 0, len(refSecrets))
		for _, secret := range refSecrets {
			secrets = append(secrets, secret.Value.MustGetPlain())
		}
		req.LogStore.UpdateRedactorAddSecrets(secrets)
	}

	var cmd []string
	if command.Script != "" {
		encodedScript := base64.StdEncoding.EncodeToString(reflectutil.UnsafeStrToBytes(command.Script))
		tmpFilePath := fmt.Sprintf("/tmp/localpaas_job_%s.sh", req.Task.ID)

		// Sample command format constructed below:
		// sh -c "echo '<base64>' | base64 -d > script-file && chmod +x script-file && script-file; exit_code=$?; \
		// rm -f script-file; exit $exit_code"
		var sb strings.Builder
		sb.Grow(len(encodedScript) + len(tmpFilePath)*5 + 100) //nolint:mnd
		sb.WriteString("echo '")
		sb.WriteString(encodedScript)
		sb.WriteString("' | base64 -d > ")
		sb.WriteString(tmpFilePath)
		sb.WriteString(" && chmod +x ")
		sb.WriteString(tmpFilePath)
		sb.WriteString(" && ")
		sb.WriteString(tmpFilePath)
		sb.WriteString("; exit_code=$?; rm -f ")
		sb.WriteString(tmpFilePath)
		sb.WriteString("; exit $exit_code")

		cmd = []string{"sh", "-c", sb.String()}
	} else {
		cmd, err = executil.CmdSplit(command.Command)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
	}

	_, err = s.ContainerExec(ctx, &containerexecservice.ContainerExecReq{
		Project:                req.Project,
		App:                    req.App,
		TaskMinRunningDuration: req.TaskMinRunningDuration,
		TaskFindRetryMax:       req.TaskFindRetryMax,
		TaskFindRetryDelay:     req.TaskFindRetryDelay,
		LogStore:               req.LogStore,
		ExecOptions: func(opts *client.ExecCreateOptions) {
			opts.AttachStdout = true
			opts.AttachStderr = true
			opts.Cmd = cmd
			opts.WorkingDir = command.WorkingDir
			opts.Env = env
			opts.TTY = command.TTY
			opts.ConsoleSize.Width = gofn.Coalesce(command.ConsoleSize.Width, docker.DefaultConsoleSize.Width)
			opts.ConsoleSize.Height = gofn.Coalesce(command.ConsoleSize.Height, docker.DefaultConsoleSize.Height)
		},
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &containerexecservice.SchedJobExecResp{}, nil
}
