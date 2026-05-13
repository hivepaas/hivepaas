package appsettingsuc

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"
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

const (
	mountKeyTypeBind    = "type:%v:src:%v:propagation:%v:target:%v:consistency:%v"
	mountKeyTypeVolume  = "type:%v:src:%v:subpath:%v:target:%v:consistency:%v"
	mountKeyTypeCluster = mountKeyTypeVolume
	mountKeyTypeTmpfs   = "type:%v:size:%v:target:%v:consistency:%v"
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
	App               *entity.App
	Service           *swarm.Service
	ExistingMountKeys map[string]struct{}
	StorageSettings   *entity.StorageSettings
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

	// Calculate mount keys of existing mounts to distinguish new changes
	existingMounts := service.Spec.TaskTemplate.ContainerSpec.Mounts
	data.ExistingMountKeys = make(map[string]struct{}, len(existingMounts))
	for i := range existingMounts {
		data.ExistingMountKeys[uc.calcExistingMountKey(&existingMounts[i])] = struct{}{}
	}

	// Load project storage settings to make sure these app settings comply with
	storageSttg, err := uc.settingRepo.GetSingle(ctx, db, base.NewSettingScopeProject(app.ProjectID),
		base.SettingTypeStorageSettings, true)
	if err != nil {
		return apperrors.Wrap(err)
	}
	data.StorageSettings = storageSttg.MustAsStorageSettings()

	for _, reqMnt := range req.Mounts {
		if !gofn.Contain(supportedMountTypes, reqMnt.Type) {
			return apperrors.NewUnsupported(fmt.Sprintf("Mount type '%v'", reqMnt.Type))
		}
		switch reqMnt.Type { //nolint:exhaustive
		case mount.TypeBind:
			err = uc.validateStorageSettingsBindMount(app, reqMnt, data)
		case mount.TypeVolume:
			err = uc.validateStorageSettingsVolumeMount(app, reqMnt, data)
		case mount.TypeCluster:
			err = uc.validateStorageSettingsClusterVolumeMount(app, reqMnt, data)
		case mount.TypeTmpfs:
			err = uc.validateStorageSettingsTmpfsMount(reqMnt, data)
		}
		if err != nil {
			return apperrors.Wrap(err)
		}
	}

	return nil
}

func (uc *UC) calcExistingMountKey(currMnt *mount.Mount) string {
	var key string
	switch currMnt.Type { //nolint:exhaustive
	case mount.TypeBind:
		propagation := ""
		if currMnt.BindOptions != nil {
			propagation = string(currMnt.BindOptions.Propagation)
		}
		key = fmt.Sprintf(mountKeyTypeBind, currMnt.Type, currMnt.Source, propagation,
			currMnt.Target, currMnt.Consistency)
	case mount.TypeVolume:
		subpath := ""
		if currMnt.VolumeOptions != nil {
			subpath = currMnt.VolumeOptions.Subpath
		}
		key = fmt.Sprintf(mountKeyTypeVolume, currMnt.Type, currMnt.Source, subpath,
			currMnt.Target, currMnt.Consistency)
	case mount.TypeCluster:
		subpath := ""
		if currMnt.VolumeOptions != nil {
			subpath = currMnt.VolumeOptions.Subpath
		}
		key = fmt.Sprintf(mountKeyTypeCluster, currMnt.Type, currMnt.Source, subpath,
			currMnt.Target, currMnt.Consistency)
	case mount.TypeTmpfs:
		size := int64(0)
		if currMnt.TmpfsOptions != nil {
			size = currMnt.TmpfsOptions.SizeBytes
		}
		key = fmt.Sprintf(mountKeyTypeTmpfs, currMnt.Type, size,
			currMnt.Target, currMnt.Consistency)
	}
	return key
}

func (uc *UC) calcRequestingMountKey(mnt *appsettingsdto.Mount) string {
	var key string
	switch mnt.Type { //nolint:exhaustive
	case mount.TypeBind:
		source := filepath.Join(mnt.BindOptions.BaseDir, mnt.BindOptions.Subpath)
		propagation := ""
		if mnt.BindOptions != nil {
			propagation = string(mnt.BindOptions.Propagation)
		}
		key = fmt.Sprintf(mountKeyTypeBind, mnt.Type, source, propagation,
			mnt.Target, mnt.Consistency)
	case mount.TypeVolume:
		volume := ""
		subpath := ""
		if mnt.VolumeOptions != nil {
			volume = mnt.VolumeOptions.Volume
			subpath = mnt.VolumeOptions.Subpath
		}
		key = fmt.Sprintf(mountKeyTypeVolume, mnt.Type, volume, subpath,
			mnt.Target, mnt.Consistency)
	case mount.TypeCluster:
		volume := ""
		subpath := ""
		if mnt.ClusterOptions != nil {
			volume = mnt.ClusterOptions.Volume
			subpath = mnt.ClusterOptions.Subpath
		}
		key = fmt.Sprintf(mountKeyTypeCluster, mnt.Type, volume, subpath,
			mnt.Target, mnt.Consistency)
	case mount.TypeTmpfs:
		size := int64(0)
		if mnt.TmpfsOptions != nil {
			size = mnt.TmpfsOptions.Size.Bytes()
		}
		key = fmt.Sprintf(mountKeyTypeTmpfs, mnt.Type, size,
			mnt.Target, mnt.Consistency)
	}
	return key
}

func (uc *UC) validateStorageSettingsBindMount(
	app *entity.App,
	mnt *appsettingsdto.Mount,
	data *updateAppStorageSettingsData,
) error {
	// If the requesting mount exists and no change, keep it
	if _, exists := data.ExistingMountKeys[uc.calcRequestingMountKey(mnt)]; exists {
		return nil
	}

	bindSettings := data.StorageSettings.BindSettings
	if bindSettings == nil || !bindSettings.Enabled {
		return apperrors.New(apperrors.ErrUnconfigured).WithParam("Name", "Bind settings")
	}

	if len(bindSettings.BaseDirs) > 0 {
		contain, _ := fileutil.PathContain(bindSettings.BaseDirs, mnt.BindOptions.BaseDir)
		if !contain {
			return apperrors.New(apperrors.ErrSettingViolated).
				WithParam("Name", fmt.Sprintf("Use of base dir '%v'", mnt.BindOptions.BaseDir))
		}
	}

	subpathRequired := bindSettings.CaclRequiredSubpath(app)
	if subpathRequired != "" {
		isSubpath, _ := fileutil.IsEqualOrSubpath(subpathRequired, mnt.BindOptions.Subpath)
		if !isSubpath {
			return apperrors.New(apperrors.ErrSettingViolated).
				WithParam("Name", fmt.Sprintf("Use of subpath '%v'", mnt.BindOptions.Subpath))
		}
	}

	return nil
}

func (uc *UC) validateStorageSettingsVolumeMount(
	app *entity.App,
	mnt *appsettingsdto.Mount,
	data *updateAppStorageSettingsData,
) error {
	// If the requesting mount exists and no change, keep it
	if _, exists := data.ExistingMountKeys[uc.calcRequestingMountKey(mnt)]; exists {
		return nil
	}

	volumeSettings := data.StorageSettings.VolumeSettings
	if volumeSettings == nil || !volumeSettings.Enabled {
		return apperrors.New(apperrors.ErrUnconfigured).WithParam("Name", "Volume settings")
	}

	if len(volumeSettings.Volumes) > 0 &&
		!gofn.Contain(volumeSettings.Volumes.ToIDStringSlice(), mnt.VolumeOptions.Volume) {
		return apperrors.New(apperrors.ErrSettingViolated).
			WithParam("Name", fmt.Sprintf("Use of volume '%v'", mnt.VolumeOptions.Volume))
	}

	subpathRequired := volumeSettings.CaclRequiredSubpath(app)
	if subpathRequired != "" {
		isSubpath, _ := fileutil.IsEqualOrSubpath(subpathRequired, mnt.VolumeOptions.Subpath)
		if !isSubpath {
			return apperrors.New(apperrors.ErrSettingViolated).
				WithParam("Name", fmt.Sprintf("Use of subpath '%v'", mnt.VolumeOptions.Subpath))
		}
	}

	return nil
}

func (uc *UC) validateStorageSettingsClusterVolumeMount(
	app *entity.App,
	mnt *appsettingsdto.Mount,
	data *updateAppStorageSettingsData,
) error {
	// If the requesting mount exists and no change, keep it
	if _, exists := data.ExistingMountKeys[uc.calcRequestingMountKey(mnt)]; exists {
		return nil
	}

	volumeSettings := data.StorageSettings.ClusterVolumeSettings
	if volumeSettings == nil || !volumeSettings.Enabled {
		return apperrors.New(apperrors.ErrUnconfigured).WithParam("Name", "Cluster volume settings")
	}

	if len(volumeSettings.Volumes) > 0 &&
		!gofn.Contain(volumeSettings.Volumes.ToIDStringSlice(), mnt.ClusterOptions.Volume) {
		return apperrors.New(apperrors.ErrSettingViolated).
			WithParam("Name", fmt.Sprintf("Use of volume '%v'", mnt.ClusterOptions.Volume))
	}

	subpathRequired := volumeSettings.CaclRequiredSubpath(app)
	if subpathRequired != "" {
		isSubpath, _ := fileutil.IsEqualOrSubpath(subpathRequired, mnt.ClusterOptions.Subpath)
		if !isSubpath {
			return apperrors.New(apperrors.ErrSettingViolated).
				WithParam("Name", fmt.Sprintf("Use of subpath '%v'", mnt.ClusterOptions.Subpath))
		}
	}

	return nil
}

func (uc *UC) validateStorageSettingsTmpfsMount(
	mnt *appsettingsdto.Mount,
	data *updateAppStorageSettingsData,
) error {
	// If the requesting mount exists and no change, keep it
	if _, exists := data.ExistingMountKeys[uc.calcRequestingMountKey(mnt)]; exists {
		return nil
	}

	tmpfsSettings := data.StorageSettings.TmpfsSettings
	if tmpfsSettings == nil || !tmpfsSettings.Enabled {
		return apperrors.New(apperrors.ErrUnconfigured).WithParam("Name", "Tmpfs settings")
	}

	var size int64
	if mnt.TmpfsOptions != nil {
		size = int64(mnt.TmpfsOptions.Size)
	}
	if tmpfsSettings.MaxSize > 0 && size > int64(tmpfsSettings.MaxSize) {
		return apperrors.New(apperrors.ErrSettingViolated).
			WithParam("Name", fmt.Sprintf("Tmpfs size '%v'", size))
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
		// Use type and source to identify a mount
		var subpath string
		if mnt.VolumeOptions != nil {
			subpath = mnt.VolumeOptions.Subpath
		}
		currMountMap[fmt.Sprintf("type:%v:src:%v:subpath:%v", mnt.Type, mnt.Source, subpath)] = mnt
	}

	for _, reqMnt := range req.Mounts {
		var source, subpath string
		switch reqMnt.Type { //nolint:exhaustive
		case mount.TypeBind:
			source = filepath.Join(reqMnt.BindOptions.BaseDir, reqMnt.BindOptions.Subpath)
		case mount.TypeVolume:
			source = reqMnt.VolumeOptions.Volume
			subpath = reqMnt.VolumeOptions.Subpath
		case mount.TypeCluster:
			source = reqMnt.ClusterOptions.Volume
			subpath = reqMnt.ClusterOptions.Subpath
		}
		key := fmt.Sprintf("type:%v:src:%v:subpath:%v", reqMnt.Type, source, subpath)

		mnt := currMountMap[key]
		if mnt == nil {
			mnt = &mount.Mount{
				Type:   reqMnt.Type,
				Source: source,
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
				mnt.VolumeOptions.Subpath = subpath
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
				mnt.VolumeOptions.Subpath = subpath
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
