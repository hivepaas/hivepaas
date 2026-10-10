package appsettingsuc

import (
	"context"
	"slices"
	"strconv"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
	"github.com/hivepaas/hivepaas/services/docker"
	"github.com/hivepaas/hivepaas/services/docker/dockerhelper"
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
			return hperrors.Wrap(err)
		}

		err = uc.applyAppResourceSettings(ctx, req, data)
		if err != nil {
			return hperrors.Wrap(err)
		}

		return uc.recordAppUpdate(ctx, db, auth, data.App, base.AuditLogSourceAPIUpdate, "resources", nil)
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
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

	if data.Service == nil || appsettingsdto.ResourceSettingsVersion(data.Service) != req.UpdateVer {
		return hperrors.Wrap(hperrors.ErrUpdateVerMismatched)
	}

	// Modifying capabilities requires Write on Cluster module, as does giving
	// the app a GPU or taking it away - by Enable GPU or by a reservation.
	currCaps := appsettingsdto.TransformCapabilities(&service.Spec.TaskTemplate)
	if !req.Capabilities.Equal(currCaps) || reservesGPU(req, &service.Spec.TaskTemplate) !=
		dockerhelper.ReservesGPU(&service.Spec.TaskTemplate) {
		hasPerm, err := uc.permissionManager.CheckAccess(ctx, db, auth, &permission.ModuleAccessCheck{
			BaseAccessCheck: permission.BaseAccessCheck{Action: base.ActionTypeWrite},
			Module:          base.ResourceModuleCluster,
		})
		if err != nil {
			return hperrors.Wrap(err)
		}
		if !hasPerm {
			return hperrors.Wrap(hperrors.ErrUnauthorized).WithMsgLog(
				"changing capabilities requires Write permission on Cluster module")
		}
	}

	return nil
}

func (uc *UC) prepareUpdatingAppResourceSettings(
	req *appsettingsdto.UpdateAppResourceSettingsReq,
	data *updateAppResourceSettingsData,
) {
	// Enable GPU is a reservation the screen shows apart from the others: a
	// request saying nothing of the capabilities keeps it.
	enableGPU := dockerhelper.ReservesOneGPU(&data.Service.Spec.TaskTemplate)
	if req.Capabilities != nil {
		enableGPU = req.Capabilities.EnableGPU
	}
	uc.prepareUpdatingAppResourceReservations(req, data)
	uc.prepareUpdatingAppResourceLimits(req, data)
	uc.prepareUpdatingAppMemory(req, data)
	uc.prepareUpdatingAppCapabilities(req, data)
	dockerhelper.SetOneGPU(&data.Service.Spec.TaskTemplate, enableGPU)
}

// reservesGPU says whether the request leaves the app a GPU: by Enable GPU, by
// a reservation, or - saying nothing of either - as it has now.
func reservesGPU(req *appsettingsdto.UpdateAppResourceSettingsReq, task *swarm.TaskSpec) bool {
	if req.Capabilities != nil && req.Capabilities.EnableGPU {
		return true
	}
	if req.Reservations != nil && slices.ContainsFunc(req.Reservations.GenericResources,
		func(r *appsettingsdto.GenericResource) bool { return r != nil && r.Kind == docker.GenericResourceGPU }) {
		return true
	}
	return req.Capabilities == nil && dockerhelper.ReservesOneGPU(task)
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

	// What the settings leave out the app does not have, as with its limits: a
	// field emptied is docker's default again, not the last value it had.
	memory := req.Memory
	if memory == nil {
		memory = &appsettingsdto.Memory{}
	}

	taskSpec.Resources.SwapBytes = nil
	if memory.Swap != nil {
		taskSpec.Resources.SwapBytes = new(memory.Swap.Truncate(unit.MB).Bytes())
	}

	taskSpec.Resources.MemorySwappiness = memory.Swappiness

	if memory.ShmSize != nil && *memory.ShmSize > unit.MB {
		dockerhelper.SetShmSize(taskSpec, memory.ShmSize.Truncate(unit.MB).Bytes())
	} else {
		dockerhelper.RemoveShmMount(taskSpec)
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

	containerSpec.Ulimits = make([]*container.Ulimit, 0, len(req.Capabilities.Ulimits))
	for _, limit := range req.Capabilities.Ulimits {
		if limit == nil {
			continue
		}
		containerSpec.Ulimits = append(containerSpec.Ulimits, &container.Ulimit{
			Name: limit.Name,
			Hard: limit.Hard,
			Soft: limit.Soft,
		})
	}

	containerSpec.CapabilityAdd = req.Capabilities.CapabilityAdd
	containerSpec.CapabilityDrop = req.Capabilities.CapabilityDrop

	containerSpec.OomScoreAdj = req.Capabilities.OomScoreAdj
	containerSpec.Sysctls = req.Capabilities.Sysctls
}

func (uc *UC) applyAppResourceSettings(
	ctx context.Context,
	req *appsettingsdto.UpdateAppResourceSettingsReq,
	data *updateAppResourceSettingsData,
) error {
	// What the screen showed, checked again on a retry: swarm refuses an update
	// of a service it wrote meanwhile, and the retry reads the service again.
	shown := appsettingsdto.ResourceSettingsVersion(data.Service)
	err := uc.dockerManager.ServiceUpdateFunc(ctx, data.Service.ID, data.Service,
		func(attempt int, service *swarm.Service) (bool, error) {
			if attempt > 0 && appsettingsdto.ResourceSettingsVersion(service) != shown {
				return false, hperrors.Wrap(hperrors.ErrUpdateVerMismatched)
			}
			data.Service = service
			uc.prepareUpdatingAppResourceSettings(req, data)
			return true, nil
		}, defaultServiceRetryMax, 0)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
