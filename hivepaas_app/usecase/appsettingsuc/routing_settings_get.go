package appsettingsuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/entityutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/settinghelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
)

func (uc *UC) GetAppRoutingSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *appsettingsdto.GetAppRoutingSettingsReq,
) (*appsettingsdto.GetAppRoutingSettingsResp, error) {
	app, err := uc.appService.LoadApp(ctx, uc.db, req.ProjectID, req.AppID, false, false,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
		bunex.SelectRelation("ProjectEnv"),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	settings, _, err := uc.settingRepo.List(ctx, uc.db, nil, nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeAppRouting),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
		bunex.SelectWhere("setting.object_id = ?", app.ID),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	input := &appsettingsdto.AppRoutingSettingsTransformInput{
		App:             app,
		RoutingSettings: settinghelper.FindSettingByType(settings, base.SettingTypeAppRouting),
	}

	err = uc.loadAppRoutingSettingsRefData(ctx, uc.db, input)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	resp, err := appsettingsdto.TransformRoutingSettings(input)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &appsettingsdto.GetAppRoutingSettingsResp{
		Data: resp,
	}, nil
}

func (uc *UC) loadAppRoutingSettingsRefData(
	ctx context.Context,
	db database.IDB,
	input *appsettingsdto.AppRoutingSettingsTransformInput,
) (err error) {
	if input.RoutingSettings == nil {
		return nil
	}

	app := input.App
	routingSettings, err := input.RoutingSettings.AsAppRoutingSettings()
	if err != nil {
		return hperrors.Wrap(err)
	}
	settingIDs := routingSettings.GetRefObjectIDs().RefSettingIDs

	settings, _, err := uc.settingRepo.List(ctx, db, app.GetObjectScope(), nil,
		bunex.SelectWhere("setting.id IN (?)", bunex.List(settingIDs)),
	)
	if err != nil {
		return hperrors.Wrap(err)
	}
	for _, setting := range settings {
		setting.CurrentObjectID = app.ID
	}
	input.RefSettingMap = entityutil.SliceToIDMap(settings)

	return nil
}
