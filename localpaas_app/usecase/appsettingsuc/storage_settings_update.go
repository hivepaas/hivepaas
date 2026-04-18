package appsettingsuc

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/fileutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/transaction"
	"github.com/localpaas/localpaas/localpaas_app/usecase/appsettingsuc/appsettingsdto"
)

var (
	supportedMountTypes = []mount.Type{mount.TypeBind, mount.TypeVolume, mount.TypeCluster, mount.TypeTmpfs}
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

		persistingData := &persistingAppData{}
		uc.prepareUpdatingAppStorageSettings(req, data)

		err = uc.persistData(ctx, db, persistingData)
		if err != nil {
			return apperrors.Wrap(err)
		}

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
	App     *entity.App
	Service *swarm.Service
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

	// Load project storage settings to make sure these app settings comply with
	storageSttg, err := uc.settingRepo.GetSingle(ctx, db, base.NewSettingScopeProject(app.ProjectID),
		base.SettingTypeStorageSettings, true)
	if err != nil {
		return apperrors.Wrap(err)
	}
	storageSettings := storageSttg.MustAsStorageSettings()

	for _, reqMnt := range req.Mounts {
		if !gofn.Contain(supportedMountTypes, reqMnt.Type) {
			return apperrors.NewUnsupported(fmt.Sprintf("Mount type '%v'", reqMnt.Type))
		}
		switch reqMnt.Type { //nolint:exhaustive
		case mount.TypeBind:
			err = uc.validateStorageSettingsBindMount(app, reqMnt, storageSettings)
		case mount.TypeVolume:
			err = uc.validateStorageSettingsVolumeMount(app, reqMnt, storageSettings)
		case mount.TypeCluster:
			err = uc.validateStorageSettingsClusterVolumeMount(app, reqMnt, storageSettings)
		case mount.TypeTmpfs:
			err = uc.validateStorageSettingsTmpfsMount(reqMnt, storageSettings)
		}
		if err != nil {
			return apperrors.Wrap(err)
		}
	}

	return nil
}

func (uc *UC) validateStorageSettingsBindMount(
	app *entity.App,
	mnt *appsettingsdto.Mount,
	storageSettings *entity.StorageSettings,
) error {
	bindSettings := storageSettings.BindSettings
	if bindSettings == nil || !bindSettings.Enabled {
		return apperrors.NewUnavailable("Bind settings is not configured")
	}
	if len(bindSettings.BaseDirs) == 0 {
		return nil
	}

	var appSubpath string
	if bindSettings.AppsMustUseSubPaths {
		appSubpath = filepath.Join(bindSettings.BaseSubpath, app.Project.Key, app.LocalKey)
	}

	for _, baseDir := range bindSettings.BaseDirs {
		if bindSettings.AppsMustUseSubPaths {
			baseDir = filepath.Join(baseDir, appSubpath)
		}
		isSubpath, err := fileutil.IsEqualOrSubpath(baseDir, mnt.Source)
		if err != nil {
			return apperrors.Wrap(err)
		}
		if isSubpath {
			return nil
		}
	}
	return apperrors.NewParamInvalid(fmt.Sprintf("Bind source '%v'", mnt.Source))
}

func (uc *UC) validateStorageSettingsVolumeMount(
	app *entity.App,
	mnt *appsettingsdto.Mount,
	storageSettings *entity.StorageSettings,
) error {
	volumeSettings := storageSettings.VolumeSettings
	if volumeSettings == nil || !volumeSettings.Enabled {
		return apperrors.NewUnavailable("Volume settings is not configured")
	}
	if len(volumeSettings.Volumes) == 0 {
		return nil
	}

	if !gofn.Contain(volumeSettings.Volumes.ToIDStringSlice(), mnt.Source) {
		return apperrors.NewParamInvalid(fmt.Sprintf("Volume '%v'", mnt.Source))
	}

	if volumeSettings.AppsMustUseSubPaths {
		appSubpath := filepath.Join(volumeSettings.BaseSubpath, app.Project.Key, app.LocalKey)
		isSubpath, err := fileutil.IsEqualOrSubpath(appSubpath, mnt.VolumeOptions.Subpath)
		if err != nil {
			return apperrors.Wrap(err)
		}
		if isSubpath {
			return nil
		}
	}

	return apperrors.NewParamInvalid(fmt.Sprintf("Volume '%v'", mnt.Source))
}

func (uc *UC) validateStorageSettingsClusterVolumeMount(
	app *entity.App,
	mnt *appsettingsdto.Mount,
	storageSettings *entity.StorageSettings,
) error {
	volumeSettings := storageSettings.ClusterVolumeSettings
	if volumeSettings == nil || !volumeSettings.Enabled {
		return apperrors.NewUnavailable("Cluster volume settings is not configured")
	}
	if len(volumeSettings.Volumes) == 0 {
		return nil
	}

	if !gofn.Contain(volumeSettings.Volumes.ToIDStringSlice(), mnt.Source) {
		return apperrors.NewParamInvalid(fmt.Sprintf("Cluster volume '%v'", mnt.Source))
	}

	if volumeSettings.AppsMustUseSubPaths {
		appSubpath := filepath.Join(volumeSettings.BaseSubpath, app.Project.Key, app.LocalKey)
		isSubpath, err := fileutil.IsEqualOrSubpath(appSubpath, mnt.VolumeOptions.Subpath)
		if err != nil {
			return apperrors.Wrap(err)
		}
		if isSubpath {
			return nil
		}
	}

	return apperrors.NewParamInvalid(fmt.Sprintf("Cluster volume '%v'", mnt.Source))
}

func (uc *UC) validateStorageSettingsTmpfsMount(
	mnt *appsettingsdto.Mount,
	storageSettings *entity.StorageSettings,
) error {
	tmpfsSettings := storageSettings.TmpfsSettings
	if tmpfsSettings == nil || !tmpfsSettings.Enabled {
		return apperrors.NewUnavailable("Tmpfs settings is not configured")
	}

	var size int64
	if mnt.TmpfsOptions != nil {
		size = int64(mnt.TmpfsOptions.Size)
	}
	if tmpfsSettings.MaxSize > 0 && size > int64(tmpfsSettings.MaxSize) {
		return apperrors.NewParamInvalid(fmt.Sprintf("Tmpfs size '%v'", size))
	}

	return nil
}

func (uc *UC) prepareUpdatingAppStorageSettings(
	req *appsettingsdto.UpdateAppStorageSettingsReq,
	data *updateAppStorageSettingsData,
) {
	uc.prepareUpdatingAppStorageMounts(req, data)
}

//nolint:gocognit
func (uc *UC) prepareUpdatingAppStorageMounts(
	req *appsettingsdto.UpdateAppStorageSettingsReq,
	data *updateAppStorageSettingsData,
) {
	service := data.Service
	containerSpec := service.Spec.TaskTemplate.ContainerSpec

	if len(req.Mounts) == 0 {
		containerSpec.Mounts = nil
		return
	}

	currMounts := containerSpec.Mounts
	containerSpec.Mounts = make([]mount.Mount, 0, len(req.Mounts))

	currMountMap := make(map[string]*mount.Mount, len(containerSpec.Mounts))
	for i := range currMounts {
		mnt := &currMounts[i]
		if !gofn.Contain(supportedMountTypes, mnt.Type) {
			// Keep the unsupported mounts
			containerSpec.Mounts = append(containerSpec.Mounts, *mnt)
			continue
		}
		// Use type and source to identify a mount (add subpath if Volume mount)
		key := fmt.Sprintf("type:%v:src:%v", mnt.Type, mnt.Source)
		if mnt.Type == mount.TypeVolume && mnt.VolumeOptions != nil && mnt.VolumeOptions.Subpath != "" {
			key += fmt.Sprintf(":subpath:%v", mnt.VolumeOptions.Subpath)
		}
		currMountMap[key] = mnt
	}

	for _, reqMnt := range req.Mounts {
		key := fmt.Sprintf("type:%v:src:%v", reqMnt.Type, reqMnt.Source)
		if reqMnt.Type == mount.TypeVolume && reqMnt.VolumeOptions != nil && reqMnt.VolumeOptions.Subpath != "" {
			key += fmt.Sprintf(":subpath:%v", reqMnt.VolumeOptions.Subpath)
		}

		mnt := currMountMap[key]
		if mnt == nil {
			mnt = &mount.Mount{
				Type:   reqMnt.Type,
				Source: reqMnt.Source,
			}
		}

		mnt.Target = reqMnt.Target
		mnt.Consistency = reqMnt.Consistency
		switch reqMnt.Type { //nolint:exhaustive
		case mount.TypeBind:
			if mnt.BindOptions == nil {
				mnt.BindOptions = &mount.BindOptions{}
			}
			if reqMnt.BindOptions != nil {
				mnt.BindOptions.Propagation = reqMnt.BindOptions.Propagation
				mnt.BindOptions.NonRecursive = reqMnt.BindOptions.NonRecursive
				mnt.BindOptions.CreateMountpoint = reqMnt.BindOptions.CreateMountpoint
				mnt.BindOptions.ReadOnlyNonRecursive = reqMnt.BindOptions.ReadOnlyNonRecursive
				mnt.BindOptions.ReadOnlyForceRecursive = reqMnt.BindOptions.ReadOnlyForceRecursive
			}

		case mount.TypeVolume:
			if mnt.VolumeOptions == nil && reqMnt.VolumeOptions != nil {
				mnt.VolumeOptions = &mount.VolumeOptions{}
			}
			if reqMnt.VolumeOptions != nil {
				mnt.VolumeOptions.NoCopy = reqMnt.VolumeOptions.NoCopy
				mnt.VolumeOptions.Subpath = reqMnt.VolumeOptions.Subpath
				mnt.VolumeOptions.Labels = reqMnt.VolumeOptions.Labels
				if reqMnt.VolumeOptions.DriverConfig != nil {
					mnt.VolumeOptions.DriverConfig = &mount.Driver{
						Name:    reqMnt.VolumeOptions.DriverConfig.Name,
						Options: reqMnt.VolumeOptions.DriverConfig.Options,
					}
				} else {
					mnt.VolumeOptions.DriverConfig = nil
				}
			}

		case mount.TypeCluster:
			if mnt.VolumeOptions == nil && reqMnt.ClusterOptions != nil {
				mnt.VolumeOptions = &mount.VolumeOptions{}
			}
			if reqMnt.ClusterOptions != nil {
				mnt.VolumeOptions.NoCopy = reqMnt.ClusterOptions.NoCopy
				mnt.VolumeOptions.Subpath = reqMnt.ClusterOptions.Subpath
				mnt.VolumeOptions.Labels = reqMnt.ClusterOptions.Labels
				if reqMnt.ClusterOptions.DriverConfig != nil {
					mnt.VolumeOptions.DriverConfig = &mount.Driver{
						Name:    reqMnt.ClusterOptions.DriverConfig.Name,
						Options: reqMnt.ClusterOptions.DriverConfig.Options,
					}
				} else {
					mnt.VolumeOptions.DriverConfig = nil
				}
			}

		case mount.TypeTmpfs:
			if mnt.TmpfsOptions == nil {
				mnt.TmpfsOptions = &mount.TmpfsOptions{}
			}
			mnt.TmpfsOptions.SizeBytes = reqMnt.TmpfsOptions.Size.Bytes()
			mnt.TmpfsOptions.Mode = reqMnt.TmpfsOptions.Mode.ToFileMode()
			mnt.TmpfsOptions.Options = reqMnt.TmpfsOptions.Options
		}

		containerSpec.Mounts = append(containerSpec.Mounts, *mnt)
	}
}

func (uc *UC) applyAppStorageSettings(
	ctx context.Context,
	data *updateAppStorageSettingsData,
) error {
	service := data.Service

	_, err := uc.dockerManager.ServiceUpdate(ctx, service.ID, &service.Version, &service.Spec)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}
