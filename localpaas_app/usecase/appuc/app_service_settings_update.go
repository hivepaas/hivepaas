package appuc

import (
	"context"

	"github.com/docker/docker/api/types/swarm"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/transaction"
	"github.com/localpaas/localpaas/localpaas_app/usecase/appuc/appdto"
	"github.com/localpaas/localpaas/services/docker"
)

func (uc *UC) UpdateAppServiceSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *appdto.UpdateAppServiceSettingsReq,
) (*appdto.UpdateAppServiceSettingsResp, error) {
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		data := &updateAppServiceSettingsData{}
		err := uc.loadAppServiceSettingsForUpdate(ctx, db, req, data)
		if err != nil {
			return apperrors.Wrap(err)
		}

		persistingData := &persistingAppData{}
		uc.prepareUpdatingAppServiceSettings(req, data)

		err = uc.persistData(ctx, db, persistingData)
		if err != nil {
			return apperrors.Wrap(err)
		}

		err = uc.applyAppServiceSettings(ctx, data)
		if err != nil {
			return apperrors.Wrap(err)
		}
		return nil
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &appdto.UpdateAppServiceSettingsResp{}, nil
}

type updateAppServiceSettingsData struct {
	App     *entity.App
	Service *swarm.Service
}

func (uc *UC) loadAppServiceSettingsForUpdate(
	ctx context.Context,
	db database.Tx,
	req *appdto.UpdateAppServiceSettingsReq,
	data *updateAppServiceSettingsData,
) error {
	app, err := uc.appService.LoadApp(ctx, db, req.ProjectID, req.AppID, true, true,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectFor("UPDATE OF app"),
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}
	data.App = app

	service, err := uc.appService.ServiceInspect(ctx, app.ServiceID, false)
	if err != nil {
		return apperrors.Wrap(err)
	}
	data.Service = service

	if data.Service == nil || data.Service.Version.Index != uint64(req.UpdateVer) { //nolint:gosec
		return apperrors.Wrap(apperrors.ErrUpdateVerMismatched)
	}

	return nil
}

func (uc *UC) prepareUpdatingAppServiceSettings(
	req *appdto.UpdateAppServiceSettingsReq,
	data *updateAppServiceSettingsData,
) {
	uc.prepareUpdatingAppServiceMode(req, data)
	uc.prepareUpdatingAppServicePlacement(req, data)
}

func (uc *UC) prepareUpdatingAppServiceMode(
	req *appdto.UpdateAppServiceSettingsReq,
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
	req *appdto.UpdateAppServiceSettingsReq,
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

func (uc *UC) applyAppServiceSettings(
	ctx context.Context,
	data *updateAppServiceSettingsData,
) error {
	service := data.Service

	_, err := uc.dockerManager.ServiceUpdate(ctx, service.ID, &service.Version, &service.Spec)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}
