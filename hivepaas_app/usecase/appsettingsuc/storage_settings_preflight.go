package appsettingsuc

import (
	"context"

	"github.com/moby/moby/api/types/mount"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
)

// PreflightAppStorageSettings reports which of the mounts about to be added
// would reach a directory that already holds something.
//
// Only what is being added is asked about. A mount the app already has reaches
// the app's own data by definition, and warning about that would be warning that
// the app has been running.
func (uc *UC) PreflightAppStorageSettings(
	ctx context.Context,
	_ *basedto.Auth,
	req *appsettingsdto.PreflightAppStorageSettingsReq,
) (*appsettingsdto.PreflightAppStorageSettingsResp, error) {
	app, err := uc.appService.LoadApp(ctx, uc.db, req.ProjectID, req.AppID, false, false,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
		bunex.SelectRelation("ProjectEnv"),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	kept, err := uc.currentMountKeys(ctx, app)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	inspectReq := &volumeservice.InspectAppStorageReq{Scope: app.GetObjectScope()}
	targets := map[string]string{}
	for _, reqMnt := range req.Mounts {
		if reqMnt.Key != "" && kept[reqMnt.Key] {
			continue
		}
		// A mount that names another app reaches somebody else's directory on
		// purpose, and what is in there is not a leftover of this app.
		if reqMnt.SourceApp != nil || reqMnt.Source == "" {
			continue
		}
		if reqMnt.Type != mount.TypeVolume && reqMnt.Type != mount.TypeCluster {
			continue
		}
		key := reqMnt.Source + "|" + mountReqSubpath(reqMnt)
		targets[key] = reqMnt.Target
		inspectReq.Queries = append(inspectReq.Queries, &volumeservice.AppStorageQuery{
			AppKey:   key,
			App:      app,
			VolumeID: reqMnt.Source,
			Subpath:  mountReqSubpath(reqMnt),
		})
	}

	data := &appsettingsdto.PreflightAppStorageResult{
		Storage:          []*appsettingsdto.PreflightStorageRes{},
		StorageUnchecked: []*appsettingsdto.PreflightStorageRes{},
	}
	if len(inspectReq.Queries) == 0 {
		return &appsettingsdto.PreflightAppStorageSettingsResp{Data: data}, nil
	}

	resp, err := uc.volumeService.InspectAppStorage(ctx, uc.db, inspectReq)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	for _, state := range resp.States {
		if state.Checked && !state.HasData() {
			continue
		}
		item := &appsettingsdto.PreflightStorageRes{Target: targets[state.AppKey], Path: state.Path}
		item.Volume.ID, item.Volume.Name = state.VolumeID, state.VolumeName
		if state.Checked {
			data.Storage = append(data.Storage, item)
			continue
		}
		data.StorageUnchecked = append(data.StorageUnchecked, item)
	}
	return &appsettingsdto.PreflightAppStorageSettingsResp{Data: data}, nil
}

// currentMountKeys is what the app already mounts, by the key the screen carries
// an unchanged mount back with.
func (uc *UC) currentMountKeys(ctx context.Context, app *entity.App) (map[string]bool, error) {
	keys := map[string]bool{}
	if app.ServiceID == "" {
		return keys, nil
	}
	service, err := uc.clusterService.ServiceInspect(ctx, app.ServiceID, false)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if service == nil || service.Spec.TaskTemplate.ContainerSpec == nil {
		return keys, nil
	}
	mounts := service.Spec.TaskTemplate.ContainerSpec.Mounts
	for i := range mounts {
		keys[uc.calcMountKey(&mounts[i])] = true
	}
	return keys, nil
}

func mountReqSubpath(mnt *appsettingsdto.Mount) string {
	switch {
	case mnt.VolumeOptions != nil:
		return mnt.VolumeOptions.Subpath
	case mnt.ClusterOptions != nil:
		return mnt.ClusterOptions.Subpath
	default:
		return ""
	}
}
