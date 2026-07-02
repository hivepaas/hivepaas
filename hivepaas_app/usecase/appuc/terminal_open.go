package appuc

import (
	"context"

	"github.com/moby/moby/client"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/containerexecservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appuc/appdto"
	"github.com/hivepaas/hivepaas/services/docker"
)

func (uc *UC) OpenTerminal(
	ctx context.Context,
	auth *basedto.Auth,
	req *appdto.OpenTerminalReq,
) (_ *appdto.OpenTerminalResp, err error) {
	app, featureSettings, err := uc.appService.LoadAppWithFeatureSettings(ctx, uc.db, req.ProjectID, req.AppID,
		true, true,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
	)
	if err != nil {
		return nil, apperrors.New(err)
	}
	if app.ServiceID == "" {
		return nil, apperrors.NewUnavailable("App service").
			WithMsgLog("service not exist for app")
	}
	if featureSettings.TerminalSettings != nil && !featureSettings.TerminalSettings.Enabled {
		return nil, apperrors.NewUnavailable("App terminal")
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
		return nil, apperrors.New(err)
	}

	return &appdto.OpenTerminalResp{
		ExecAttachResult: execResp.ExecAttachResult,
		ExecResizeFunc:   execResp.ExecResizeFunc,
		CloseFunc:        execResp.CloseFunc,
	}, nil
}
