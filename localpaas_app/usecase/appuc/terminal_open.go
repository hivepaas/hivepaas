package appuc

import (
	"context"
	"errors"

	"github.com/moby/moby/client"
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/service/containerexecservice"
	"github.com/localpaas/localpaas/localpaas_app/usecase/appuc/appdto"
	"github.com/localpaas/localpaas/services/docker"
)

func (uc *UC) OpenTerminal(
	ctx context.Context,
	auth *basedto.Auth,
	req *appdto.OpenTerminalReq,
) (_ *appdto.OpenTerminalResp, err error) {
	app, err := uc.appService.LoadApp(ctx, uc.db, req.ProjectID, req.AppID, true, true,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
	)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	if app.ServiceID == "" {
		return nil, apperrors.New(apperrors.ErrUnavailable).
			WithMsgLog("service not exist for app")
	}

	terminalEnabled := true
	featureSetting, err := uc.settingRepo.GetSingle(ctx, uc.db, app.GetSettingScope(),
		base.SettingTypeAppFeatures, true)
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return nil, apperrors.Wrap(err)
	}
	if featureSetting != nil {
		featureSettings := featureSetting.MustAsAppFeatureSettings()
		if featureSettings.TerminalSettings != nil {
			terminalEnabled = featureSettings.TerminalSettings.Enabled
		}
	}
	if !terminalEnabled {
		return nil, apperrors.NewForbidden("Terminal")
	}

	execResp, err := uc.containerExecService.ContainerExec(ctx, &containerexecservice.ContainerExecReq{
		Project:      app.Project,
		App:          app,
		TerminalMode: true,
		ExecOptions: func(opts *client.ExecCreateOptions) {
			opts.AttachStdin = true
			opts.AttachStdout = true
			opts.AttachStderr = true
			opts.TTY = true
			opts.Cmd = []string{gofn.Coalesce(req.Shell, "sh")}
			opts.ConsoleSize.Width = gofn.Coalesce(req.Width, docker.DefaultConsoleSize.Width)
			opts.ConsoleSize.Height = gofn.Coalesce(req.Height, docker.DefaultConsoleSize.Height)
		},
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &appdto.OpenTerminalResp{
		ExecAttachResult: execResp.ExecAttachResult,
		ExecResizeFunc:   execResp.ExecResizeFunc,
		CloseFunc:        execResp.Close,
	}, nil
}
