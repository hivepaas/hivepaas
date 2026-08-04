package appuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appuc/appdto"
)

func (uc *UC) PrepareAppClone(
	ctx context.Context,
	auth *basedto.Auth,
	req *appdto.PrepareAppCloneReq,
) (*appdto.PrepareAppCloneResp, error) {
	app, err := uc.appService.LoadApp(ctx, uc.db, req.ProjectID, req.AppID, true, false,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
		bunex.SelectRelation("ProjectEnv"),
		bunex.SelectRelation("Settings",
			bunex.SelectWhere("setting.type = ?", base.SettingTypeAppHttp),
		),
	)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	refObjects, err := uc.settingService.LoadReferenceObjects(ctx, uc.db, app.GetObjectScope(),
		true, false, app.Settings...)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp, err := appdto.TransformAppClonePreparationData(app, refObjects)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &appdto.PrepareAppCloneResp{
		Data: resp,
	}, nil
}
