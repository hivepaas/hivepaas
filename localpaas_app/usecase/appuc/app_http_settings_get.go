package appuc

import (
	"context"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/entityutil"
	"github.com/localpaas/localpaas/localpaas_app/usecase/appuc/appdto"
)

func (uc *UC) GetAppHttpSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *appdto.GetAppHttpSettingsReq,
) (*appdto.GetAppHttpSettingsResp, error) {
	app, err := uc.appRepo.GetByID(ctx, uc.db, req.ProjectID, req.AppID,
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

	input := &appdto.AppHttpSettingsTransformInput{
		App:          app,
		HttpSettings: gofn.FirstOr(settings, nil),
	}

	err = uc.loadAppHttpSettingsRefData(ctx, uc.db, input)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp, err := appdto.TransformHttpSettings(input)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &appdto.GetAppHttpSettingsResp{
		Data: resp,
	}, nil
}

func (uc *UC) loadAppHttpSettingsRefData(
	ctx context.Context,
	db database.IDB,
	input *appdto.AppHttpSettingsTransformInput,
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

	settings, _, err := uc.settingRepo.List(ctx, db, app.GetSettingScope(), nil,
		bunex.SelectWhere("setting.id IN (?)", bunex.In(settingIDs)),
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
