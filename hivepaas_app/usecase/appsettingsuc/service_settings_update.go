package appsettingsuc

import (
	"context"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/copier"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
	"github.com/hivepaas/hivepaas/services/docker"
)

func (uc *UC) UpdateAppServiceSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *appsettingsdto.UpdateAppServiceSettingsReq,
) (*appsettingsdto.UpdateAppServiceSettingsResp, error) {
	data := &updateAppServiceSettingsData{}
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		err := uc.loadAppServiceSettingsForUpdate(ctx, db, req, data)
		if err != nil {
			return hperrors.Wrap(err)
		}

		err = uc.applyAppServiceSettings(ctx, db, req, data)
		if err != nil {
			return hperrors.Wrap(err)
		}
		return nil
	})
	if err != nil {
		// A failed mode change rolls the swarm service back, but under a new ID. The transaction is
		// rolled back too, so that ID has to be stored outside of it or the app would keep pointing
		// at the service that was deleted.
		if data.RestoredServiceID != "" && data.App != nil {
			_ = uc.persistAppServiceID(ctx, uc.db, data.App, data.RestoredServiceID)
		}
		return nil, hperrors.Wrap(err)
	}

	return &appsettingsdto.UpdateAppServiceSettingsResp{}, nil
}

type updateAppServiceSettingsData struct {
	App     *entity.App
	Service *swarm.Service

	// RestoredServiceID is set when a mode change failed and the previous service had to be
	// recreated, which gives it a new ID that must be persisted outside the rolled back
	// transaction.
	RestoredServiceID string
}

func (uc *UC) loadAppServiceSettingsForUpdate(
	ctx context.Context,
	db database.Tx,
	req *appsettingsdto.UpdateAppServiceSettingsReq,
	data *updateAppServiceSettingsData,
) error {
	app, err := uc.appService.LoadApp(ctx, db, req.ProjectID, req.AppID, true, true,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectFor("UPDATE OF app"),
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
		bunex.SelectRelation("ProjectEnv"),
	)
	if err != nil {
		return hperrors.Wrap(err)
	}
	data.App = app

	service, err := uc.clusterService.ServiceInspect(ctx, app.ServiceID, false)
	if err != nil {
		return hperrors.Wrap(err)
	}
	data.Service = service

	if data.Service == nil || data.Service.Version.Index != uint64(req.UpdateVer) { //nolint:gosec
		return hperrors.Wrap(hperrors.ErrUpdateVerMismatched)
	}

	return nil
}

func (uc *UC) prepareUpdatingAppServiceSettings(
	req *appsettingsdto.UpdateAppServiceSettingsReq,
	data *updateAppServiceSettingsData,
) {
	uc.prepareUpdatingAppServiceMode(req, data)
	uc.prepareUpdatingAppServicePlacement(req, data)
}

func (uc *UC) prepareUpdatingAppServiceMode(
	req *appsettingsdto.UpdateAppServiceSettingsReq,
	data *updateAppServiceSettingsData,
) {
	service := data.Service
	spec := &service.Spec
	currMode := &spec.Mode
	spec.Mode = swarm.ServiceMode{}
	switch req.ModeSpec.Mode {
	case docker.ServiceModeReplicated:
		item := currMode.Replicated
		if item == nil {
			item = &swarm.ReplicatedService{}
		}
		item.Replicas = req.ModeSpec.ServiceReplicas
		spec.Mode.Replicated = item
	case docker.ServiceModeReplicatedJob:
		item := currMode.ReplicatedJob
		if item == nil {
			item = &swarm.ReplicatedJob{}
		}
		item.MaxConcurrent = req.ModeSpec.JobMaxConcurrent
		item.TotalCompletions = req.ModeSpec.JobTotalCompletions
		spec.Mode.ReplicatedJob = item
	case docker.ServiceModeGlobal:
		item := currMode.Global
		if item == nil {
			item = &swarm.GlobalService{}
		}
		spec.Mode.Global = item
	case docker.ServiceModeGlobalJob:
		item := currMode.GlobalJob
		if item == nil {
			item = &swarm.GlobalJob{}
		}
		spec.Mode.GlobalJob = item
	}
}

func (uc *UC) prepareUpdatingAppServicePlacement(
	req *appsettingsdto.UpdateAppServiceSettingsReq,
	data *updateAppServiceSettingsData,
) {
	service := data.Service
	taskSpec := &service.Spec.TaskTemplate
	if req.Placement == nil {
		taskSpec.Placement = nil
		return
	}

	if taskSpec.Placement == nil {
		taskSpec.Placement = &swarm.Placement{}
	}

	taskSpec.Placement.Constraints = make([]string, 0, len(req.Placement.Constraints))
	for _, constraint := range req.Placement.Constraints {
		taskSpec.Placement.Constraints = append(taskSpec.Placement.Constraints,
			constraint.Name+constraint.Op+constraint.Value)
	}

	taskSpec.Placement.Preferences = make([]swarm.PlacementPreference, 0, len(req.Placement.Preferences))
	for _, pref := range req.Placement.Preferences {
		if pref.Name == "spread" {
			taskSpec.Placement.Preferences = append(taskSpec.Placement.Preferences,
				swarm.PlacementPreference{
					Spread: &swarm.SpreadOver{SpreadDescriptor: pref.Value},
				})
		}
	}
}

// isChangingServiceMode reports whether the request switches the service to another mode variant
// (replicated <-> global <-> jobs). Swarm rejects such a change on an existing service, so it can
// only be applied by recreating the service.
func isChangingServiceMode(
	req *appsettingsdto.UpdateAppServiceSettingsReq,
	service *swarm.Service,
) bool {
	if req.ModeSpec == nil || service == nil {
		return false
	}
	currMode := appsettingsdto.TransformServiceMode(&service.Spec)
	if currMode == nil || currMode.Mode == "" {
		return false
	}
	return currMode.Mode != req.ModeSpec.Mode
}

// isAppStopped reports whether the service is a replicated one scaled down to 0, which is how
// HivePaaS represents a stopped app.
func isAppStopped(service *swarm.Service) bool {
	if service == nil || service.Spec.Mode.Replicated == nil {
		return false
	}
	replicas := service.Spec.Mode.Replicated.Replicas
	return replicas == nil || *replicas == 0
}

func (uc *UC) applyAppServiceSettings(
	ctx context.Context,
	db database.Tx,
	req *appsettingsdto.UpdateAppServiceSettingsReq,
	data *updateAppServiceSettingsData,
) error {
	if isChangingServiceMode(req, data.Service) {
		if isAppStopped(data.Service) {
			return hperrors.Wrap(hperrors.ErrServiceModeChangeRequiresRunningApp)
		}
		return uc.recreateAppServiceWithNewMode(ctx, db, req, data)
	}

	err := uc.dockerManager.ServiceUpdateFunc(ctx, data.Service.ID, data.Service,
		func(_ int, service *swarm.Service) (bool, error) {
			data.Service = service
			uc.prepareUpdatingAppServiceSettings(req, data)
			return true, nil
		}, defaultServiceRetryMax, 0)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

// recreateAppServiceWithNewMode applies the settings by deleting and recreating the swarm service,
// which is the only way to switch its mode variant. This stops the app while it runs.
func (uc *UC) recreateAppServiceWithNewMode(
	ctx context.Context,
	db database.Tx,
	req *appsettingsdto.UpdateAppServiceSettingsReq,
	data *updateAppServiceSettingsData,
) error {
	// Keep an untouched copy to restore if the new service cannot be created. It must be a deep
	// copy: the spec shares its Labels map and Placement pointer with the live service, so the
	// updates below would otherwise leak into the copy meant for the rollback.
	oldSpec, err := copier.CopyAs(data.Service.Spec)
	if err != nil {
		return hperrors.Wrap(err)
	}

	uc.prepareUpdatingAppServiceSettings(req, data)
	newSpec := data.Service.Spec

	// The stored mode belongs to the mode being replaced, so starting the app later must not
	// restore it.
	delete(newSpec.Labels, appservice.LabelAppPrevServiceMode)

	previousServiceID := data.App.ServiceID

	newServiceID, err := uc.appService.RecreateServiceWithSpec(ctx, data.App, &oldSpec, &newSpec)
	if err != nil {
		// A rollback recreated the previous service under a new ID. Hand it to the caller: this
		// transaction is about to be rolled back, so it cannot be persisted here.
		if data.App.ServiceID != previousServiceID {
			data.RestoredServiceID = data.App.ServiceID
		}
		return hperrors.Wrap(err)
	}

	return uc.persistAppServiceID(ctx, db, data.App, newServiceID)
}

func (uc *UC) persistAppServiceID(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	serviceID string,
) error {
	app.ServiceID = serviceID
	app.UpdateVer++
	app.UpdatedAt = timeutil.NowUTC()

	err := uc.appRepo.Update(ctx, db, app,
		bunex.UpdateColumns("service_id", "update_ver", "updated_at"))
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
