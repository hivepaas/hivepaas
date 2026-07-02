package appsettingsuc

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
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
)

func (uc *UC) GetAppHttpSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *appsettingsdto.GetAppHttpSettingsReq,
) (*appsettingsdto.GetAppHttpSettingsResp, error) {
	app, err := uc.appRepo.GetByID(ctx, uc.db, req.ProjectID, req.AppID,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
	)
	if err != nil {
		return nil, apperrors.New(err)
	}

	settings, _, err := uc.settingRepo.List(ctx, uc.db, nil, nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeAppHttp),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
		bunex.SelectWhere("setting.object_id = ?", app.ID),
	)
	if err != nil {
		return nil, apperrors.New(err)
	}

	input := &appsettingsdto.AppHttpSettingsTransformInput{
		App:          app,
		HttpSettings: gofn.FirstOr(settings, nil),
	}

	err = uc.loadAppHttpSettingsRefData(ctx, uc.db, input)
	if err != nil {
		return nil, apperrors.New(err)
	}

	resp, err := appsettingsdto.TransformHttpSettings(input)
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &appsettingsdto.GetAppHttpSettingsResp{
		Data: resp,
	}, nil
}

func (uc *UC) loadAppHttpSettingsRefData(
	ctx context.Context,
	db database.IDB,
	input *appsettingsdto.AppHttpSettingsTransformInput,
) (err error) {
	if input.HttpSettings == nil {
		return nil
	}

	app := input.App
	appHttpSettings, err := input.HttpSettings.AsAppHttpSettings()
	if err != nil {
		return apperrors.New(err)
	}
	settingIDs := appHttpSettings.GetRefObjectIDs().RefSettingIDs

	settings, _, err := uc.settingRepo.List(ctx, db, app.GetObjectScope(), nil,
		bunex.SelectWhere("setting.id IN (?)", bunex.List(settingIDs)),
	)
	if err != nil {
		return apperrors.New(err)
	}
	for _, setting := range settings {
		setting.CurrentObjectID = app.ID
	}
	input.RefSettingMap = entityutil.SliceToIDMap(settings)

	return nil
}
