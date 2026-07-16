package hpappsettingsuc

import (
	"context"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/entityutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/hpappsettingsuc/hpappsettingsdto"
)

func (uc *UC) GetHttpSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *hpappsettingsdto.GetHttpSettingsReq,
) (*hpappsettingsdto.GetHttpSettingsResp, error) {
	app, err := uc.appRepo.GetByKey(ctx, uc.db, "", base.HivepaasAppKey,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
	)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	settings, _, err := uc.settingRepo.List(ctx, uc.db, nil, nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeAppHttp),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
		bunex.SelectWhere("setting.object_id = ?", app.ID),
	)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	input := &hpappsettingsdto.HttpSettingsTransformInput{
		App:          app,
		HttpSettings: gofn.FirstOr(settings, nil),
	}

	err = uc.loadHttpSettingsRefData(ctx, uc.db, input)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp, err := hpappsettingsdto.TransformHttpSettings(input)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &hpappsettingsdto.GetHttpSettingsResp{
		Data: resp,
	}, nil
}

func (uc *UC) loadHttpSettingsRefData(
	ctx context.Context,
	db database.IDB,
	input *hpappsettingsdto.HttpSettingsTransformInput,
) (err error) {
	if input.HttpSettings == nil {
		return nil
	}

	app := input.App
	appHttpSettings, err := input.HttpSettings.AsAppHttpSettings()
	if err != nil {
		return apperrors.Wrap(err)
	}
	settingIDs := appHttpSettings.GetRefObjectIDs().RefSettingIDs

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
