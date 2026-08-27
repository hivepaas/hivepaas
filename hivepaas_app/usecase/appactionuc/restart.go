package appactionuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appactionuc/appactiondto"
)

func (uc *UC) RestartApp(
	ctx context.Context,
	auth *basedto.Auth,
	req *appactiondto.RestartAppReq,
) (*appactiondto.RestartAppResp, error) {
	app, err := uc.appService.LoadApp(ctx, uc.db, req.ProjectID, req.AppID, true, true,
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
		bunex.SelectRelation("ProjectEnv"),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	err = uc.dockerManager.ServiceForceUpdate(ctx, app.ServiceID)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &appactiondto.RestartAppResp{}, nil
}
