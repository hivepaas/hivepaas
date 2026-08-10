package appsettingsuc

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/api/types/volume"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/dockerhelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/entityutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
	"github.com/hivepaas/hivepaas/services/docker"
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
			return apperrors.Wrap(err)
		}

		uc.prepareUpdatingAppStorageSettings(ctx, data)

		err = uc.applyAppStorageSettings(ctx, data)
		if err != nil {
			return apperrors.Wrap(err)
		}
		return nil
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &appsettingsdto.UpdateAppStorageSettingsResp{}, nil
}

type updateAppStorageSettingsData struct {
	App           *entity.App
	Service       *swarm.Service
	DBVolumes     map[string]*entity.Setting
	DockerVolumes map[string]*volume.Volume
	FinalMounts   []mount.Mount
	NewMountReqs  []*appsettingsdto.Mount
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
		return apperrors.Wrap(err)
	}
	data.App = app

	service, err := uc.clusterService.ServiceInspect(ctx, app.ServiceID, false)
	if err != nil {
		return apperrors.Wrap(err)
	}
	data.Service = service

	if data.Service == nil || data.Service.Version.Index != uint64(req.UpdateVer) { //nolint:gosec
		return apperrors.Wrap(apperrors.ErrUpdateVerMismatched)
	}

	// Calculate mount keys of existing mounts to distinguish new changes
	currMounts := service.Spec.TaskTemplate.ContainerSpec.Mounts
	mapCurrMountByKey := make(map[string]*mount.Mount, len(currMounts))
	for i := range currMounts {
		mnt := &currMounts[i]
		mapCurrMountByKey[uc.calcMountKey(mnt)] = mnt
	}

	var newDBVolIDs []string
	var newDockerVolIDs []string
	for _, reqMnt := range req.Mounts {
		if existingMount, exists := mapCurrMountByKey[reqMnt.Key]; reqMnt.Key != "" && exists {
			data.FinalMounts = append(data.FinalMounts, *existingMount) // unchanged mount
			continue
		}

		// For custom mounts, only support type Volume and Cluster
		if reqMnt.Type != mount.TypeVolume && reqMnt.Type != mount.TypeCluster {
			return apperrors.Wrap(apperrors.ErrUnsupported).
				WithParam("Name", fmt.Sprintf("Mount type '%v'", reqMnt.Type))
		}
		dbVolID := reqMnt.Source
		data.NewMountReqs = append(data.NewMountReqs, reqMnt)
		newDBVolIDs = append(newDBVolIDs, dbVolID)
	}

	// Validate volumes can be used by the project
	dbVols, err := uc.settingRepo.ListByIDs(ctx, db, app.GetObjectScope(), newDBVolIDs, true,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeClusterVolume),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}
	dbVolMap := entityutil.SliceToIDMap(dbVols)
	for _, dbVolID := range newDBVolIDs {
		dbVol, ok := dbVolMap[dbVolID]
		if !ok {
			return apperrors.NewNotFound("Volume").WithMsgLog("volume %v not found", dbVolID)
		}
		newDockerVolIDs = append(newDockerVolIDs, dbVol.MustAsClusterVolume().RefID)
	}
	data.DBVolumes = dbVolMap

	// Loads docker volumes
	listRes, err := uc.dockerManager.VolumeListByIDs(ctx, newDockerVolIDs)
	if err != nil {
		return apperrors.Wrap(err)
	}
	data.DockerVolumes = make(map[string]*volume.Volume, len(dbVolMap))
	for i := range listRes.Items {
		dockerVol := &listRes.Items[i]
		data.DockerVolumes[dockerhelper.GetVolumeID(dockerVol)] = dockerVol
	}

	return nil
}

func (uc *UC) prepareUpdatingAppStorageSettings(
	ctx context.Context,
	data *updateAppStorageSettingsData,
) {
	for _, reqMnt := range data.NewMountReqs {
		dbVol := data.DBVolumes[reqMnt.Source]
		dockerVol := data.DockerVolumes[dbVol.MustAsClusterVolume().RefID]
		dockerMnt := &mount.Mount{
			Type:        reqMnt.Type,
			Source:      dockerhelper.GetVolumeID(dockerVol),
			Target:      reqMnt.Target,
			ReadOnly:    reqMnt.ReadOnly,
			Consistency: reqMnt.Consistency,
		}

		uc.buildDockerMount(ctx, dockerMnt, reqMnt, dockerVol, dbVol, data)

		// Ensure full permissions (0777) on all mounted volume subpaths before starting/updating the service
		if dockerMnt.Type == mount.TypeVolume || dockerMnt.Type == mount.TypeCluster {
			subpath := ""
			if dockerMnt.VolumeOptions != nil {
				subpath = dockerMnt.VolumeOptions.Subpath
			}
			_ = uc.volumeService.EnsureVolumePermissions(ctx, dockerMnt, subpath)
		}

		data.FinalMounts = append(data.FinalMounts, *dockerMnt)
	}
}

func (uc *UC) buildDockerMount(
	ctx context.Context,
	dockerMnt *mount.Mount,
	reqMnt *appsettingsdto.Mount,
	dockerVol *volume.Volume,
	dbVol *entity.Setting,
	data *updateAppStorageSettingsData,
) {
	app := data.App
	subpath := calcMountSubpath(app, reqMnt, dbVol)
	uc.useBindMountIfAppropriate(ctx, dockerMnt, dockerVol, subpath)

	switch dockerMnt.Type {
	case mount.TypeVolume:
		if reqMnt.VolumeOptions != nil {
			dockerMnt.VolumeOptions = &mount.VolumeOptions{
				Subpath: subpath,
				NoCopy:  reqMnt.VolumeOptions.NoCopy,
				Labels:  reqMnt.VolumeOptions.Labels,
			}
			if reqMnt.VolumeOptions.DriverConfig != nil {
				dockerMnt.VolumeOptions.DriverConfig = &mount.Driver{
					Name:    reqMnt.VolumeOptions.DriverConfig.Name,
					Options: reqMnt.VolumeOptions.DriverConfig.Options,
				}
			}
		}
	case mount.TypeCluster:
		if reqMnt.ClusterOptions != nil {
			dockerMnt.VolumeOptions = &mount.VolumeOptions{
				Subpath: subpath,
				NoCopy:  reqMnt.ClusterOptions.NoCopy,
				Labels:  reqMnt.ClusterOptions.Labels,
			}
			if reqMnt.ClusterOptions.DriverConfig != nil {
				dockerMnt.VolumeOptions.DriverConfig = &mount.Driver{
					Name:    reqMnt.ClusterOptions.DriverConfig.Name,
					Options: reqMnt.ClusterOptions.DriverConfig.Options,
				}
			}
		}
	case mount.TypeBind, mount.TypeImage, mount.TypeTmpfs, mount.TypeNamedPipe:
	}
}

func (uc *UC) useBindMountIfAppropriate(
	ctx context.Context,
	dockerMnt *mount.Mount,
	dockerVol *volume.Volume,
	subpath string,
) {
	if dockerVol.Driver != string(docker.VolumeDriverLocal) { // not a driver local
		return
	}
	directory := dockerVol.Options["device"]
	if dockerVol.Options["type"] != "none" || directory == "" {
		return
	}

	err := uc.volumeService.MakeSubDirInHost(ctx, directory, subpath, true)
	if err != nil {
		return
	}

	dockerMnt.Type = mount.TypeBind
	dockerMnt.Source = filepath.Join(directory, subpath)
	dockerMnt.BindOptions = &mount.BindOptions{
		CreateMountpoint: true,
	}
	if propagation := getConfiguredPropagation(dockerVol.Options["o"]); propagation != "" {
		dockerMnt.BindOptions.Propagation = propagation
	}
	// Reset all other kind of options
	dockerMnt.VolumeOptions = nil
	dockerMnt.ClusterOptions = nil
	dockerMnt.TmpfsOptions = nil
	dockerMnt.ImageOptions = nil
}

func (uc *UC) applyAppStorageSettings(
	ctx context.Context,
	data *updateAppStorageSettingsData,
) error {
	inspect, err := uc.dockerManager.ServiceInspect(ctx, data.Service.ID)
	if err != nil {
		return apperrors.Wrap(err)
	}
	service := &inspect.Service
	service.Spec.TaskTemplate.ContainerSpec.Mounts = data.FinalMounts

	_, err = uc.dockerManager.ServiceUpdate(ctx, service.ID, &service.Version, &service.Spec)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}

func calcMountSubpath(
	app *entity.App,
	reqMnt *appsettingsdto.Mount,
	dbVol *entity.Setting,
) string {
	var subpath string
	switch dbVol.Scope {
	case base.ObjectScopeGlobal:
		subpath = fmt.Sprintf("%v/%v/%v", app.Project.Key, app.ProjectEnv.Key, app.Key)
	case base.ObjectScopeProject:
		subpath = fmt.Sprintf("%v/%v", app.ProjectEnv.Key, app.Key)
	case base.ObjectScopeProjectEnv, base.ObjectScopeApp:
		subpath = app.Key
	case base.ObjectScopeUser, base.ObjectScopeHivepaas:
	}

	if reqMnt.Type == mount.TypeVolume && reqMnt.VolumeOptions != nil {
		subpath = filepath.Join(subpath, reqMnt.VolumeOptions.Subpath)
	}
	if reqMnt.Type == mount.TypeCluster && reqMnt.ClusterOptions != nil {
		subpath = filepath.Join(subpath, reqMnt.ClusterOptions.Subpath)
	}

	return subpath
}

func getConfiguredPropagation(o string) mount.Propagation {
	parts := strings.Split(o, ",")
	for _, part := range parts {
		switch mount.Propagation(part) {
		case mount.PropagationRPrivate:
		case mount.PropagationPrivate:
		case mount.PropagationRSlave:
		case mount.PropagationSlave:
		case mount.PropagationRShared:
		case mount.PropagationShared:
			return mount.Propagation(part)
		}
	}
	return "" // to use default one
}
