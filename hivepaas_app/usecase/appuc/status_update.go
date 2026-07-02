package appuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appuc/appdto"
)

func (uc *UC) UpdateAppStatus(
	ctx context.Context,
	auth *basedto.Auth,
	req *appdto.UpdateAppStatusReq,
) (*appdto.UpdateAppStatusResp, error) {
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		appData := &updateAppData{}
		err := uc.loadAppDataForUpdateStatus(ctx, db, req, appData)
		if err != nil {
			return apperrors.New(err)
		}
		if !appData.HasChanges {
			return nil
		}

		err = uc.appService.SetAppStatus(ctx, db, appData.App, req.Status, true)
		if err != nil {
			return apperrors.New(err)
		}
		return nil
	})
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &appdto.UpdateAppStatusResp{}, nil
}

func (uc *UC) loadAppDataForUpdateStatus(
	ctx context.Context,
	db database.IDB,
	req *appdto.UpdateAppStatusReq,
	data *updateAppData,
) error {
	app, err := uc.appService.LoadApp(ctx, db, req.ProjectID, req.ID, true, false,
		bunex.SelectFor("UPDATE OF app"),
		bunex.SelectRelation("Project"),
	)
	if err != nil {
		return apperrors.New(err)
	}
	if app.UpdateVer != req.UpdateVer {
		return apperrors.New(apperrors.ErrUpdateVerMismatched)
	}
	data.App = app
	data.HasChanges = app.Status != req.Status

	return nil
}
