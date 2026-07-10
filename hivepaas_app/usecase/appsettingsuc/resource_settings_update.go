package appsettingsuc

import (
	"context"
	"strconv"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/dockerhelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
	"github.com/hivepaas/hivepaas/services/docker"
)

func (uc *UC) UpdateAppResourceSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *appsettingsdto.UpdateAppResourceSettingsReq,
) (*appsettingsdto.UpdateAppResourceSettingsResp, error) {
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		data := &updateAppResourceSettingsData{}
		err := uc.loadAppResourceSettingsForUpdate(ctx, db, auth, req, data)
		if err != nil {
			return apperrors.New(err)
		}

		persistingData := &persistingAppData{}
		uc.prepareUpdatingAppResourceSettings(req, data)

		err = uc.persistData(ctx, db, persistingData)
		if err != nil {
			return apperrors.New(err)
		}

		err = uc.applyAppResourceSettings(ctx, data)
		if err != nil {
			return apperrors.New(err)
		}
		return nil
	})
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &appsettingsdto.UpdateAppResourceSettingsResp{}, nil
}

type updateAppResourceSettingsData struct {
	App     *entity.App
	Service *swarm.Service
}

func (uc *UC) loadAppResourceSettingsForUpdate(
	ctx context.Context,
	db database.Tx,
	auth *basedto.Auth,
	req *appsettingsdto.UpdateAppResourceSettingsReq,
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
		return apperrors.New(err)
	}
	data.App = app

	service, err := uc.clusterService.ServiceInspect(ctx, app.ServiceID, false)
	if err != nil {
		return apperrors.New(err)
	}
	data.Service = service

	if data.Service == nil || data.Service.Version.Index != uint64(req.UpdateVer) { //nolint:gosec
		return apperrors.New(apperrors.ErrUpdateVerMismatched)
	}

	currCaps := appsettingsdto.TransformCapabilities(service.Spec.TaskTemplate.ContainerSpec)
	if !req.Capabilities.Equal(currCaps) { // Modifying capabilities requires Write on Cluster module
		hasPerm, err := uc.permissionManager.CheckAccess(ctx, db, auth, &permission.AccessCheck{
			ResourceModule: base.ResourceModuleCluster,
			Action:         base.ActionTypeWrite,
		})
		if err != nil {
			return apperrors.New(err)
		}
		if !hasPerm {
			return apperrors.New(apperrors.ErrUnauthorized).WithMsgLog(
				"changing capabilities requires Write permission on Cluster module")
		}
	}

	return nil
}

func (uc *UC) prepareUpdatingAppResourceSettings(
	req *appsettingsdto.UpdateAppResourceSettingsReq,
	data *updateAppResourceSettingsData,
) {
	uc.prepareUpdatingAppResourceReservations(req, data)
	uc.prepareUpdatingAppResourceLimits(req, data)
	uc.prepareUpdatingAppMemory(req, data)
	uc.prepareUpdatingAppResourceUlimits(req, data)
	uc.prepareUpdatingAppCapabilities(req, data)
}

func (uc *UC) prepareUpdatingAppResourceReservations(
	req *appsettingsdto.UpdateAppResourceSettingsReq,
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
	reservations.NanoCPUs = docker.TruncateCPUsAsNano(req.Reservations.CPUs, docker.MinCPUFraction)
	reservations.MemoryBytes = req.Reservations.Memory.Truncate(unit.MB).Bytes()
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
	req *appsettingsdto.UpdateAppResourceSettingsReq,
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
	limits.NanoCPUs = docker.TruncateCPUsAsNano(req.Limits.CPUs, docker.MinCPUFraction)
	limits.MemoryBytes = req.Limits.Memory.Truncate(unit.MB).Bytes()
	limits.Pids = req.Limits.Pids
}

func (uc *UC) prepareUpdatingAppMemory(
	req *appsettingsdto.UpdateAppResourceSettingsReq,
	data *updateAppResourceSettingsData,
) {
	service := data.Service
	taskSpec := &service.Spec.TaskTemplate
	if taskSpec.Resources == nil {
		taskSpec.Resources = &swarm.ResourceRequirements{}
	}

	if req.Memory != nil {
		taskSpec.Resources.SwapBytes = new(req.Memory.Swap.Truncate(unit.MB).Bytes())
		taskSpec.Resources.MemorySwappiness = req.Memory.Swappiness

		if req.Memory.ShmSize > unit.MB {
			dockerhelper.SetShmSize(taskSpec, req.Memory.ShmSize.Truncate(unit.MB).Bytes())
		}
	}
}

func (uc *UC) prepareUpdatingAppResourceUlimits(
	req *appsettingsdto.UpdateAppResourceSettingsReq,
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
	req *appsettingsdto.UpdateAppResourceSettingsReq,
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
		return apperrors.New(err)
	}

	return nil
}
