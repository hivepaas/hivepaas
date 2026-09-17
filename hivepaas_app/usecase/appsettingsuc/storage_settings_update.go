package appsettingsuc

import (
	"context"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/auditdetail"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/placementservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
)

func (uc *UC) UpdateAppStorageSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *appsettingsdto.UpdateAppStorageSettingsReq,
) (*appsettingsdto.UpdateAppStorageSettingsResp, error) {
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		data := &updateAppStorageSettingsData{}
		err := uc.loadAppStorageSettingsForUpdate(ctx, db, req, data)
		if err != nil {
			return hperrors.Wrap(err)
		}

		// Building the mounts also refuses a set that pins the service to more
		// than one node, which has to happen before the mounts reach the service.
		built, err := uc.volumeService.BuildAppMounts(ctx, db, &volumeservice.BuildAppMountsReq{
			App:  data.App,
			Kept: data.KeptMounts,
			New:  data.NewMounts,
		})
		if err != nil {
			return hperrors.Wrap(err)
		}
		data.FinalMounts = built.Mounts

		err = uc.applyAppStorageSettings(ctx, db, data)
		if err != nil {
			return hperrors.Wrap(err)
		}

		return uc.recordAppUpdate(ctx, db, auth, data.App, base.AuditLogSourceAPIUpdate, "storage", auditdetail.New().
			Set("mountCount", len(data.FinalMounts)))
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &appsettingsdto.UpdateAppStorageSettingsResp{}, nil
}

type updateAppStorageSettingsData struct {
	App         *entity.App
	Service     *swarm.Service
	KeptMounts  []mount.Mount
	NewMounts   []*volumeservice.AppMountReq
	FinalMounts []mount.Mount
}

func (uc *UC) loadAppStorageSettingsForUpdate(
	ctx context.Context,
	db database.Tx,
	req *appsettingsdto.UpdateAppStorageSettingsReq,
	data *updateAppStorageSettingsData,
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

	// Calculate mount keys of existing mounts to distinguish new changes
	currMounts := service.Spec.TaskTemplate.ContainerSpec.Mounts
	mapCurrMountByKey := make(map[string]*mount.Mount, len(currMounts))
	for i := range currMounts {
		mnt := &currMounts[i]
		mapCurrMountByKey[uc.calcMountKey(mnt)] = mnt
	}

	for _, reqMnt := range req.Mounts {
		if existingMount, exists := mapCurrMountByKey[reqMnt.Key]; reqMnt.Key != "" && exists {
			data.KeptMounts = append(data.KeptMounts, *existingMount) // unchanged mount
			continue
		}
		data.NewMounts = append(data.NewMounts, toAppMountReq(reqMnt))
	}
	return nil
}

// toAppMountReq carries a requested mount to volumeservice, which cannot take
// the DTO.
func toAppMountReq(reqMnt *appsettingsdto.Mount) *volumeservice.AppMountReq {
	out := &volumeservice.AppMountReq{
		Type:        reqMnt.Type,
		Source:      reqMnt.Source,
		Target:      reqMnt.Target,
		ReadOnly:    reqMnt.ReadOnly,
		Consistency: reqMnt.Consistency,
	}
	if opts := reqMnt.VolumeOptions; opts != nil {
		out.VolumeOptions = toAppMountVolumeOptions(opts)
	}
	if opts := reqMnt.ClusterOptions; opts != nil {
		out.ClusterOptions = toAppMountVolumeOptions(&opts.VolumeOptions)
	}
	return out
}

func toAppMountVolumeOptions(opts *appsettingsdto.VolumeOptions) *volumeservice.AppMountVolumeOptions {
	out := &volumeservice.AppMountVolumeOptions{
		Subpath: opts.Subpath,
		NoCopy:  opts.NoCopy,
		Labels:  opts.Labels,
	}
	if driver := opts.DriverConfig; driver != nil {
		out.DriverConfig = &mount.Driver{Name: driver.Name, Options: driver.Options}
	}
	return out
}

// applyAppStorageSettings writes the new mounts and the placement constraints
// they imply in a single service update.
//
// Writing the mounts rolls the service's tasks, so the volume pin has to be in
// the spec this call sends rather than in whatever spec is written next. Left to
// the next deploy, a task rescheduled by this very update can land on a node
// holding none of the data - the failure the pin exists to prevent, reached
// through the branch's own primary flow.
//
// SkipSavingToDocker is what makes one update enough: it has ApplyPlacementSettings
// mutate the spec and stop, instead of saving a second one and rolling the tasks
// again. The call sits inside the callback because a retry re-inspects the
// service and starts over from a spec carrying neither the mounts nor the
// constraint.
func (uc *UC) applyAppStorageSettings(
	ctx context.Context,
	db database.IDB,
	data *updateAppStorageSettingsData,
) error {
	err := uc.dockerManager.ServiceUpdateFunc(ctx, data.Service.ID, data.Service,
		func(_ int, service *swarm.Service) (bool, error) {
			service.Spec.TaskTemplate.ContainerSpec.Mounts = data.FinalMounts

			// The pins are resolved from the mounts just written, so this reads the
			// storage settings being saved and not the ones being replaced.
			_, err := uc.placementService.ApplyPlacementSettings(ctx, db,
				&placementservice.ApplyPlacementSettingsReq{
					App:                data.App,
					Service:            service,
					SkipSavingToDocker: true,
				})
			if err != nil {
				return false, hperrors.Wrap(err)
			}
			return true, nil
		}, defaultServiceRetryMax, 0)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
