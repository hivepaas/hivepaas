package appuc

import (
	"context"
	"errors"
	"strings"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appuc/appdto"
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
			return apperrors.New(err)
		}
		if !appData.HasChanges {
			return nil
		}

		persistingData := &persistingAppData{}
		uc.preparePersistingAppUpdate(req, appData, persistingData)

		return uc.persistData(ctx, db, persistingData)
	})
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &appdto.UpdateAppResp{}, nil
}

type updateAppData struct {
	App         *entity.App
	ServiceSpec *swarm.ServiceSpec
	HasChanges  bool
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
		return apperrors.New(err)
	}
	if app.UpdateVer != req.UpdateVer {
		return apperrors.New(apperrors.ErrUpdateVerMismatched)
	}
	data.App = app

	// To update app status, use a separate API, so we don't update it
	req.Status = app.Status

	// If name changes, need to verify its uniqueness
	if !strings.EqualFold(req.Name, app.Name) {
		conflictApp, err := uc.appRepo.GetByName(ctx, db, req.ProjectID, req.Name, bunex.SelectColumns("id"))
		if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
			return apperrors.New(err)
		}
		if conflictApp != nil {
			return apperrors.NewAlreadyExist("App").
				WithMsgLog("app name '%s' already exists", req.Name)
		}
	}

	data.HasChanges = true
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

	// NOTE: we don't allow to change app env after creation
	req.Env = app.Env

	uc.preparePersistingAppBase(app, req.AppBaseReq, timeNow, persistingData)
	persistingData.AppsToDeleteTags = append(persistingData.AppsToDeleteTags, app.ID)
	uc.preparePersistingAppTags(app, req.Tags, 0, persistingData)
}
