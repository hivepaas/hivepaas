package appsettingsuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/commandservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
)

func (uc *UC) BuildCommandTemplate(
	ctx context.Context,
	auth *basedto.Auth,
	req *appsettingsdto.BuildCommandTemplateReq,
) (*appsettingsdto.BuildCommandTemplateResp, error) {
	app, err := uc.appService.LoadApp(ctx, uc.db, req.ProjectID, req.AppID, true, true,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
		bunex.SelectRelation("ProjectEnv"),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	appScope := app.GetObjectScope()
	cmdSetting, err := uc.settingRepo.GetByID(ctx, uc.db, appScope, base.SettingTypeCommandTemplate,
		req.CommandID, true)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	cmdTemplate := cmdSetting.MustAsCommandTemplate()

	resp, err := uc.commandService.BuildCommand(ctx, uc.db, &commandservice.BuildCommandReq{
		Scope:             appScope,
		Command:           cmdTemplate,
		BuildFinalCommand: true,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &appsettingsdto.BuildCommandTemplateResp{
		Data: &appsettingsdto.CommandBuildResp{
			Command: resp.CommandString,
		},
	}, nil
}
