package appuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appcloneservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appuc/appdto"
)

func (uc *UC) CloneApp(
	ctx context.Context,
	auth *basedto.Auth,
	req *appdto.CloneAppReq,
) (_ *appdto.CloneAppResp, err error) {
	var cloneResp *appcloneservice.AppCloneResp
	defer func() {
		if cloneResp != nil && cloneResp.OnCleanup != nil { // Run the cleanup function
			_ = cloneResp.OnCleanup(err)
		}
	}()

	err = transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		data := &cloneAppData{}
		err := uc.loadAppDataForCloning(ctx, db, req, data)
		if err != nil {
			return apperrors.Wrap(err)
		}

		cloneResp, err = uc.appCloneService.CloneApp(ctx, db, &appcloneservice.AppCloneReq{
			SrcApp:        data.App,
			CloneSettings: &entity.AppCloneSettings{
				// TODO: complete this
			},
		})
		if err != nil {
			return apperrors.Wrap(err)
		}
		return nil
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &appdto.CloneAppResp{
		Data: &basedto.ObjectIDResp{ID: cloneResp.TargetApp.ID},
	}, apperrors.Wrap(err)
}

type cloneAppData struct {
	App *entity.App
}

func (uc *UC) loadAppDataForCloning(
	ctx context.Context,
	db database.IDB,
	req *appdto.CloneAppReq,
	data *cloneAppData,
) error {
	app, err := uc.appService.LoadApp(ctx, db, req.ProjectID, req.AppID, true, false,
		bunex.SelectFor("UPDATE OF app"),
		bunex.SelectRelation("Project.ProjectEnvs",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
		bunex.SelectRelation("ProjectEnv"),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}
	if app.UpdateVer != req.UpdateVer {
		return apperrors.Wrap(apperrors.ErrUpdateVerMismatched)
	}

	// TODO: clone app to another evn requires permission check

	data.App = app
	return nil
}
