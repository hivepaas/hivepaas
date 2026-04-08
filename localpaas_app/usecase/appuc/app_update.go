package appuc

import (
	"context"
	"errors"
	"strings"

	"github.com/docker/docker/api/types/swarm"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/transaction"
	"github.com/localpaas/localpaas/localpaas_app/usecase/appuc/appdto"
)

func (uc *UC) UpdateApp(
	ctx context.Context,
	auth *basedto.Auth,
	req *appdto.UpdateAppReq,
) (*appdto.UpdateAppResp, error) {
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		appData := &updateAppData{}
		err := uc.loadAppDataForUpdate(ctx, db, req, appData)
		if err != nil {
			return apperrors.Wrap(err)
		}

		persistingData := &persistingAppData{}
		uc.preparePersistingAppUpdate(req, appData, persistingData)

		return uc.persistData(ctx, db, persistingData)
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &appdto.UpdateAppResp{}, nil
}

type updateAppData struct {
	App         *entity.App
	ServiceSpec *swarm.ServiceSpec
}

func (uc *UC) loadAppDataForUpdate(
	ctx context.Context,
	db database.IDB,
	req *appdto.UpdateAppReq,
	data *updateAppData,
) error {
	app, err := uc.appService.LoadApp(ctx, db, req.ProjectID, req.ID, true, false,
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

	// If name changes, need to verify it uniqueness
	if !strings.EqualFold(req.Name, app.Name) {
		conflictApp, err := uc.appRepo.GetByName(ctx, db, req.ProjectID, req.Name, bunex.SelectColumns("id"))
		if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
			return apperrors.Wrap(err)
		}
		if conflictApp != nil {
			return apperrors.NewAlreadyExist("App").
				WithMsgLog("app name '%s' already exists", req.Name)
		}
	}

	return nil
}

func (uc *UC) preparePersistingAppUpdate(
	req *appdto.UpdateAppReq,
	data *updateAppData,
	persistingData *persistingAppData,
) {
	timeNow := timeutil.NowUTC()
	app := data.App
	app.UpdateVer++

	uc.preparePersistingAppBase(app, req.AppBaseReq, timeNow, persistingData)
	persistingData.AppsToDeleteTags = append(persistingData.AppsToDeleteTags, app.ID)
	uc.preparePersistingAppTags(app, req.Tags, 0, persistingData)
}
