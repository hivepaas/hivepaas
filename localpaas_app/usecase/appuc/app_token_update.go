package appuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/transaction"
	"github.com/localpaas/localpaas/localpaas_app/usecase/appuc/appdto"
)

func (uc *UC) UpdateAppToken(
	ctx context.Context,
	auth *basedto.Auth,
	req *appdto.UpdateAppTokenReq,
) (*appdto.UpdateAppTokenResp, error) {
	var appData *updateAppTokenData
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		appData = &updateAppTokenData{}
		err := uc.loadAppDataForUpdateToken(ctx, db, req, appData)
		if err != nil {
			return apperrors.Wrap(err)
		}
		uc.preparePersistingAppTokenUpdate(appData)
		return uc.persistAppTokenData(ctx, db, appData)
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &appdto.UpdateAppTokenResp{
		Data: &appdto.AppTokenDataResp{Token: appData.App.Token},
	}, nil
}

type updateAppTokenData struct {
	App *entity.App
}

func (uc *UC) loadAppDataForUpdateToken(
	ctx context.Context,
	db database.IDB,
	req *appdto.UpdateAppTokenReq,
	data *updateAppTokenData,
) error {
	app, err := uc.appService.LoadApp(ctx, db, req.ProjectID, req.ID, true, true,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectFor("UPDATE OF app"),
		bunex.SelectRelation("Project"),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}
	if app.UpdateVer != req.UpdateVer {
		return apperrors.Wrap(apperrors.ErrUpdateVerMismatched)
	}
	data.App = app

	return nil
}

func (uc *UC) preparePersistingAppTokenUpdate(
	data *updateAppTokenData,
) {
	data.App.ResetToken()
	data.App.UpdateVer++
	data.App.UpdatedAt = timeutil.NowUTC()
}

func (uc *UC) persistAppTokenData(
	ctx context.Context,
	db database.IDB,
	data *updateAppTokenData,
) error {
	err := uc.appRepo.Update(ctx, db, data.App,
		bunex.UpdateColumns("token", "update_ver", "updated_at"),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}
