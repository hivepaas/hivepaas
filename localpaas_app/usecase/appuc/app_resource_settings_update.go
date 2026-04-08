package appuc

import (
	"context"
	"strconv"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/transaction"
	"github.com/localpaas/localpaas/localpaas_app/usecase/appuc/appdto"
	"github.com/localpaas/localpaas/services/docker"
)

func (uc *UC) UpdateAppResourceSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *appdto.UpdateAppResourceSettingsReq,
) (*appdto.UpdateAppResourceSettingsResp, error) {
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		data := &updateAppResourceSettingsData{}
		err := uc.loadAppResourceSettingsForUpdate(ctx, db, req, data)
		if err != nil {
			return apperrors.Wrap(err)
		}

		persistingData := &persistingAppData{}
		uc.prepareUpdatingAppResourceSettings(req, data)

		err = uc.persistData(ctx, db, persistingData)
		if err != nil {
			return apperrors.Wrap(err)
		}

		err = uc.applyAppResourceSettings(ctx, data)
		if err != nil {
			return apperrors.Wrap(err)
		}
		return nil
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &appdto.UpdateAppResourceSettingsResp{}, nil
}

type updateAppResourceSettingsData struct {
	App     *entity.App
	Service *swarm.Service
}

func (uc *UC) loadAppResourceSettingsForUpdate(
	ctx context.Context,
	db database.Tx,
	req *appdto.UpdateAppResourceSettingsReq,
	data *updateAppResourceSettingsData,
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

func (uc *UC) prepareUpdatingAppResourceSettings(
	req *appdto.UpdateAppResourceSettingsReq,
	data *updateAppResourceSettingsData,
) {
	uc.prepareUpdatingAppResourceReservations(req, data)
	uc.prepareUpdatingAppResourceLimits(req, data)
	uc.prepareUpdatingAppResourceUlimits(req, data)
	uc.prepareUpdatingAppCapabilities(req, data)
}

func (uc *UC) prepareUpdatingAppResourceReservations(
	req *appdto.UpdateAppResourceSettingsReq,
	data *updateAppResourceSettingsData,
) {
	service := data.Service
	taskSpec := &service.Spec.TaskTemplate
	if taskSpec.Resources == nil {
		taskSpec.Resources = &swarm.ResourceRequirements{}
	}

	if req.Reservations == nil {
		taskSpec.Resources.Reservations = nil
		return
	}

	if taskSpec.Resources.Reservations == nil {
		taskSpec.Resources.Reservations = &swarm.Resources{}
	}
	reservations := taskSpec.Resources.Reservations
	reservations.NanoCPUs = int64(req.Reservations.CPUs * docker.UnitCPUNano)
	reservations.MemoryBytes = req.Reservations.MemoryMB * docker.UnitMemMB
	reservations.GenericResources = make([]swarm.GenericResource, 0, len(req.Reservations.GenericResources))

	for _, r := range req.Reservations.GenericResources {
		num, err := strconv.ParseInt(r.Value, 10, 64)
		res := swarm.GenericResource{}
		if err != nil {
			res.NamedResourceSpec = &swarm.NamedGenericResource{
				Kind:  r.Kind,
				Value: r.Value,
			}
		} else {
			res.DiscreteResourceSpec = &swarm.DiscreteGenericResource{
				Kind:  r.Kind,
				Value: num,
			}
		}
		reservations.GenericResources = append(reservations.GenericResources, res)
	}
}

func (uc *UC) prepareUpdatingAppResourceLimits(
	req *appdto.UpdateAppResourceSettingsReq,
	data *updateAppResourceSettingsData,
) {
	service := data.Service
	taskSpec := &service.Spec.TaskTemplate
	if taskSpec.Resources == nil {
		taskSpec.Resources = &swarm.ResourceRequirements{}
	}

	if req.Limits == nil {
		taskSpec.Resources.Limits = nil
		return
	}

	if taskSpec.Resources.Limits == nil {
		taskSpec.Resources.Limits = &swarm.Limit{}
	}
	limits := taskSpec.Resources.Limits
	limits.NanoCPUs = int64(req.Limits.CPUs * docker.UnitCPUNano)
	limits.MemoryBytes = req.Limits.MemoryMB * docker.UnitMemMB
	limits.Pids = req.Limits.Pids
}

func (uc *UC) prepareUpdatingAppResourceUlimits(
	req *appdto.UpdateAppResourceSettingsReq,
	data *updateAppResourceSettingsData,
) {
	service := data.Service
	containerSpec := service.Spec.TaskTemplate.ContainerSpec

	containerSpec.Ulimits = make([]*container.Ulimit, 0, len(req.Ulimits))
	for _, limit := range req.Ulimits {
		if limit == nil {
			continue
		}
		containerSpec.Ulimits = append(containerSpec.Ulimits, &container.Ulimit{
			Name: limit.Name,
			Hard: limit.Hard,
			Soft: limit.Soft,
		})
	}
}

func (uc *UC) prepareUpdatingAppCapabilities(
	req *appdto.UpdateAppResourceSettingsReq,
	data *updateAppResourceSettingsData,
) {
	if req.Capabilities == nil {
		return
	}
	service := data.Service
	containerSpec := service.Spec.TaskTemplate.ContainerSpec

	containerSpec.CapabilityAdd = req.Capabilities.CapabilityAdd
	containerSpec.CapabilityDrop = req.Capabilities.CapabilityDrop
	if req.Capabilities.EnableGPU && !gofn.Contain(containerSpec.CapabilityAdd, "[gpu]") {
		containerSpec.CapabilityAdd = append(containerSpec.CapabilityAdd, "[gpu]")
	} else if !req.Capabilities.EnableGPU {
		containerSpec.CapabilityAdd = gofn.Drop(containerSpec.CapabilityAdd, "[gpu]")
	}
	containerSpec.OomScoreAdj = req.Capabilities.OomScoreAdj
	containerSpec.Sysctls = req.Capabilities.Sysctls
}

func (uc *UC) applyAppResourceSettings(
	ctx context.Context,
	data *updateAppResourceSettingsData,
) error {
	service := data.Service

	_, err := uc.dockerManager.ServiceUpdate(ctx, service.ID, &service.Version, &service.Spec)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}
