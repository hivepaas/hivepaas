package hpappsettingsuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/entityutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/settinghelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/hpappsettingsuc/hpappsettingsdto"
)

func (uc *UC) GetRoutingSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *hpappsettingsdto.GetRoutingSettingsReq,
) (*hpappsettingsdto.GetRoutingSettingsResp, error) {
	app, err := uc.hpAppService.LoadAppByKey(ctx, uc.db, base.HivepaasAppKey,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
	)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	settings, _, err := uc.settingRepo.List(ctx, uc.db, nil, nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeAppRouting),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
		bunex.SelectWhere("setting.object_id = ?", app.ID),
	)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	input := &hpappsettingsdto.RoutingSettingsTransformInput{
		App:             app,
		RoutingSettings: settinghelper.FindSettingByType(settings, base.SettingTypeAppRouting),
	}

	err = uc.loadRoutingSettingsRefData(ctx, uc.db, input)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp, err := hpappsettingsdto.TransformRoutingSettings(input)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &hpappsettingsdto.GetRoutingSettingsResp{
		Data: resp,
	}, nil
}

func (uc *UC) loadRoutingSettingsRefData(
	ctx context.Context,
	db database.IDB,
	input *hpappsettingsdto.RoutingSettingsTransformInput,
) (err error) {
	if input.RoutingSettings == nil {
		return nil
	}

	app := input.App
	routingSettings, err := input.RoutingSettings.AsAppRoutingSettings()
	if err != nil {
		return apperrors.Wrap(err)
	}
	settingIDs := routingSettings.GetRefObjectIDs().RefSettingIDs

	settings, _, err := uc.settingRepo.List(ctx, db, app.GetObjectScope(), nil,
		bunex.SelectWhere("setting.id IN (?)", bunex.List(settingIDs)),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}
	for _, setting := range settings {
		setting.CurrentObjectID = app.ID
	}
	input.RefSettingMap = entityutil.SliceToIDMap(settings)

	return nil
}
