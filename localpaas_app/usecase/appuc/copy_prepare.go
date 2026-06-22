package appuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/usecase/appuc/appdto"
)

func (uc *UC) PrepareAppCopy(
	ctx context.Context,
	auth *basedto.Auth,
	req *appdto.PrepareAppCopyReq,
) (*appdto.PrepareAppCopyResp, error) {
	app, err := uc.appRepo.GetByID(ctx, uc.db, req.ProjectID, req.AppID,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
		bunex.SelectRelation("Settings",
			bunex.SelectWhere("setting.type = ?", base.SettingTypeAppHttp),
		),
	)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	if app.ProjectID != req.ProjectID {
		return nil, apperrors.New(apperrors.ErrUnauthorized)
	}

	refObjects, err := uc.settingService.LoadReferenceObjects(ctx, uc.db, app.GetObjectScope(),
		true, false, app.Settings...)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp, err := appdto.TransformAppCopyPreparationData(app, refObjects)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &appdto.PrepareAppCopyResp{
		Data: resp,
	}, nil
}
