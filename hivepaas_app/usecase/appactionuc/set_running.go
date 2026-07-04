package appactionuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appactionuc/appactiondto"
)

func (uc *UC) SetAppRunning(
	ctx context.Context,
	auth *basedto.Auth,
	req *appactiondto.SetAppRunningReq,
) (*appactiondto.SetAppRunningResp, error) {
	app, err := uc.appService.LoadApp(ctx, uc.db, req.ProjectID, req.AppID, true, true,
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
	)
	if err != nil {
		return nil, apperrors.New(err)
	}

	err = uc.appService.SetAppRunning(ctx, app, req.Running)
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &appactiondto.SetAppRunningResp{}, nil
}
