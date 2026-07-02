package appsettingsuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
)

func (uc *UC) GetAppEnvVars(
	ctx context.Context,
	auth *basedto.Auth,
	req *appsettingsdto.GetAppEnvVarsReq,
) (*appsettingsdto.GetAppEnvVarsResp, error) {
	app, err := uc.appRepo.GetByID(ctx, uc.db, req.ProjectID, req.AppID,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
	)
	if err != nil {
		return nil, apperrors.New(err)
	}

	settings, _, err := uc.settingRepo.List(ctx, uc.db, app.GetObjectScope(), nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeEnvVar),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
	)
	if err != nil {
		return nil, apperrors.New(err)
	}

	resp, err := appsettingsdto.TransformEnvVars(app, settings)
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &appsettingsdto.GetAppEnvVarsResp{
		Data: resp,
	}, nil
}
