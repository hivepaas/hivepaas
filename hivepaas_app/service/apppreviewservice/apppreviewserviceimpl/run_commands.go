package apppreviewserviceimpl

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/moby/moby/client"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/entityutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/executil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/reflectutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/commandservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/containerexecservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

const (
	previewCmdTaskFindRetryMax       = 5
	previewCmdTaskFindRetryDelay     = time.Second * 5
	previewCmdTaskMinRunningDuration = time.Second * 15
)

func (s *service) runCommands(
	ctx context.Context,
	db database.IDB,
	data *createPreviewData,
) (err error) {
	if data.FeatureSettings == nil || data.FeatureSettings.PreviewSettings == nil {
		return nil
	}
	cmdObjectIDs := data.FeatureSettings.PreviewSettings.Commands
	if len(cmdObjectIDs) == 0 {
		return nil
	}

	app := data.App
	if app.ServiceID == "" {
		return hperrors.Wrap(hperrors.ErrActionNotAllowed).
			WithMsgLog("parent app [%s] container is not running to execute preview commands", app.Name)
	}

	cmdIDs := cmdObjectIDs.ToIDStringSlice()
	cmdSettings, _, err := s.settingRepo.List(ctx, db, nil, nil,
		bunex.SelectWhereIn("setting.id IN (?)", cmdIDs...),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
	)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if len(cmdSettings) == 0 {
		return nil
	}

	cmdSettingMap := entityutil.SliceToIDMap(cmdSettings)

	// Ensure ref objects for command templates (e.g. referenced script files) are loaded
	err = s.settingService.LoadRefObjects(ctx, db, &data.RefObjects, app.GetObjectScope(), true, cmdSettings...)
	if err != nil {
		return hperrors.Wrap(err)
	}

	for _, cmdObj := range cmdObjectIDs {
		if cmdObj.ID == "" {
			continue
		}
		cmdSetting := cmdSettingMap[cmdObj.ID]
		if cmdSetting == nil {
			continue
		}

		err = s.runSingleCommand(ctx, db, cmdSetting, data)
		if err != nil {
			return hperrors.Wrap(err)
		}
	}

	return nil
}

func (s *service) runSingleCommand(
	ctx context.Context,
	db database.IDB,
	cmdSetting *entity.Setting,
	data *createPreviewData,
) (err error) {
	cmdTemplate := cmdSetting.MustAsCommandTemplate()

	cmd, err := s.calcCommand(ctx, cmdTemplate, data)
	if err != nil {
		return hperrors.Wrap(err)
	}

	env, err := s.calcCommandEnv(ctx, db, cmdTemplate, data)
	if err != nil {
		return hperrors.Wrap(err)
	}

	_ = data.LogStore.Add(ctx, tasklog.NewOutFrame(
		fmt.Sprintf("Executing preview command '%s' on app [%s]...", cmdSetting.Name, data.App.Name),
		tasklog.TsNow,
	))

	_, err = s.containerExecService.ContainerExec(ctx, &containerexecservice.ContainerExecReq{
		App:                    data.App,
		TaskMinRunningDuration: previewCmdTaskMinRunningDuration,
		TaskFindRetryMax:       previewCmdTaskFindRetryMax,
		TaskFindRetryDelay:     previewCmdTaskFindRetryDelay,
		LogStore:               data.LogStore,
		ExecOptions: func(opts *client.ExecCreateOptions) {
			opts.AttachStdout = true
			opts.AttachStderr = true
			opts.Cmd = cmd
			opts.WorkingDir = cmdTemplate.WorkingDir
			opts.Env = env
			opts.TTY = cmdTemplate.TTY
			opts.ConsoleSize.Width = gofn.Coalesce(cmdTemplate.ConsoleSize.Width, docker.DefaultConsoleSize.Width)
			opts.ConsoleSize.Height = gofn.Coalesce(cmdTemplate.ConsoleSize.Height, docker.DefaultConsoleSize.Height)
		},
	})
	if err != nil {
		return hperrors.Wrap(err)
	}

	_ = data.LogStore.Add(ctx, tasklog.NewOutFrame(
		fmt.Sprintf("Preview command '%s' completed successfully", cmdSetting.Name),
		tasklog.TsNow,
	))

	return nil
}

func (s *service) calcCommand(
	ctx context.Context,
	command *entity.CommandTemplate,
	data *createPreviewData,
) (cmd []string, err error) {
	if command == nil || (command.Command == "" && !command.Script.IsValid()) {
		_ = data.LogStore.Add(ctx, tasklog.NewErrFrame(
			"Execution command/script is empty, aborted", tasklog.TsNow))
		return nil, hperrors.Wrap(hperrors.ErrInternal).WithMsgLog("preview command/script is empty")
	}

	if command.Script.IsValid() {
		script := command.Script.Value
		if script == "" && command.Script.ID != "" {
			scriptSetting := data.RefObjects.RefSettings[command.Script.ID]
			if scriptSetting == nil {
				return nil, hperrors.NewNotFound("Script object")
			}
			script = scriptSetting.MustAsScript().Data
		}

		encodedScript := base64.StdEncoding.EncodeToString(reflectutil.UnsafeStrToBytes(script))
		tmpFilePath := fmt.Sprintf("/tmp/hivepaas_preview_cmd_%s.sh", data.RandSuffix)

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
			return nil, hperrors.Wrap(err)
		}
	}
	return cmd, nil
}

func (s *service) calcCommandEnv(
	ctx context.Context,
	db database.IDB,
	command *entity.CommandTemplate,
	data *createPreviewData,
) (env []string, err error) {
	resp, err := s.commandService.BuildCommand(ctx, db, &commandservice.BuildCommandReq{
		Scope:      data.App.GetObjectScope(),
		Command:    command,
		RefObjects: data.RefObjects,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	env = make([]string, 0, len(resp.EnvVars)+10) //nolint:mnd
	for _, v := range resp.EnvVars {
		env = append(env, v.ToString("="))
	}

	// Inject preview context environment variables
	extraEnvs := map[string]string{
		"HIVEPAAS_PREVIEW_APP_NAME":    data.CalcAppName,
		"HIVEPAAS_PREVIEW_SUBDOMAIN":   data.CalcSubdomain,
		"HIVEPAAS_PREVIEW_REPO_REF":    data.CalcRepoRef,
		"HIVEPAAS_PREVIEW_PULL_NUMBER": strconv.FormatUint(data.PullNumber, 10),
		"HIVEPAAS_PARENT_APP_NAME":     data.App.Name,
		"HIVEPAAS_PARENT_APP_ID":       data.App.ID,
	}
	if data.PreviewApp != nil {
		extraEnvs["HIVEPAAS_PREVIEW_APP_ID"] = data.PreviewApp.ID
	}
	for k, v := range extraEnvs {
		env = append(env, k+"="+v)
	}

	if data.LogStore != nil && len(resp.EnvVars) > 0 {
		secrets := make(map[string]struct{}, 10) //nolint:mnd
		for _, envVar := range resp.EnvVars {
			for secret := range envVar.RefSecrets {
				plainSecret, err := secret.Value.GetPlain()
				if err != nil {
					return nil, hperrors.Wrap(err)
				}
				secrets[plainSecret] = struct{}{}
			}
		}
		data.LogStore.UpdateRedactorAddSecrets(gofn.MapKeys(secrets))
	}

	return env, nil
}
