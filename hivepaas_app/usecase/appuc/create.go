package appuc

import (
	"context"
	"errors"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/auditdetail"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appprovisionservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clusterservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appuc/appdto"
)

func (uc *UC) CreateApp(
	ctx context.Context,
	auth *basedto.Auth,
	req *appdto.CreateAppReq,
) (resp *appdto.CreateAppResp, err error) {
	resp = &appdto.CreateAppResp{}
	var createdApp *entity.App

	defer func() {
		if rec := recover(); rec != nil {
			err = errors.Join(err, hperrors.ErrPanic)
		}
		if err != nil && createdApp != nil && createdApp.ServiceID != "" {
			_ = uc.clusterService.ServiceRemove(ctx, createdApp.ServiceID, clusterservice.ItemRemovalRetryMax, 0)
		}
	}()

	err = transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		provisioned, err := uc.appProvisionService.ProvisionApp(ctx, db, &appprovisionservice.ProvisionAppReq{
			ProjectID:    req.ProjectID,
			ProjectEnvID: req.ProjectEnvID,
			Name:         req.Name,
			Status:       req.Status,
			Note:         req.Note,
			Tags:         req.Tags,
		})
		if err != nil {
			return hperrors.Wrap(err)
		}
		createdApp = provisioned.App
		resp.Data = &basedto.ObjectIDResp{ID: createdApp.ID}

		// After persisting and still inside the transaction. ProvisionApp removes
		// the service itself when it fails; a record that fails here errors the
		// call and the deferred cleanup above removes it - so there is no path
		// that leaves a live app behind with nothing recorded.
		return uc.recordAppWrite(ctx, db, auth, createdApp,
			base.AuditLogTypeAppCreate, base.AuditLogSourceAPICreate, "create", auditdetail.New().
				Set("projectId", createdApp.ProjectID).
				Set("envId", createdApp.ProjectEnvID).
				Set("serviceId", createdApp.ServiceID))
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return resp, nil
}

type persistingAppData struct {
	appservice.PersistingAppData
}

func (uc *UC) preparePersistingAppBase(
	app *entity.App,
	req *appdto.AppBaseReq,
	timeNow time.Time,
	persistingData *persistingAppData,
) {
	app.Name = req.Name
	app.Status = req.Status
	app.Note = req.Note
	app.UpdatedAt = timeNow

	persistingData.UpsertingApps = append(persistingData.UpsertingApps, app)
}

func (uc *UC) preparePersistingAppTags(
	app *entity.App,
	tags []string,
	startIndex int,
	persistingData *persistingAppData,
) {
	index := startIndex
	for _, tag := range tags {
		persistingData.UpsertingTags = append(persistingData.UpsertingTags,
			&entity.Tag{
				ObjectID: app.ID,
				Tag:      tag,
				Index:    index,
			})
		index++
	}
}

func (uc *UC) persistData(
	ctx context.Context,
	db database.IDB,
	persistingData *persistingAppData,
) error {
	err := uc.appService.PersistAppData(ctx, db, &persistingData.PersistingAppData)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
