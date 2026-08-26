package volumeserviceimpl

import (
	"context"
	"errors"
	"path/filepath"

	"github.com/moby/moby/api/types/volume"
	"github.com/moby/moby/client"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/services/docker"
	"github.com/hivepaas/hivepaas/services/docker/dockerhelper"
)

const (
	fullFileMode = 0777
)

func (s *service) CreateProjectDefaultVolume(
	ctx context.Context,
	project *entity.Project,
) (_ *entity.Setting, _ *client.VolumeCreateResult, err error) {
	storagePathInHost := config.Current.Storage.BindSource
	if storagePathInHost == "" {
		return nil, nil, apperrors.Wrap(apperrors.ErrUnconfigured).
			WithParam("Name", "HP_STORAGE_BIND_SOURCE")
	}

	subpath := filepath.Join("project_data", project.Key)
	err = s.MakeSubDirInHost(ctx, storagePathInHost, subpath, true)
	if err != nil {
		return nil, nil, apperrors.Wrap(err)
	}

	driver := docker.VolumeDriverLocal
	driverOpts := map[string]string{
		"type":   "none",
		"device": filepath.Join(storagePathInHost, subpath),
		"o":      "bind,rw",
	}
	createResp, err := s.dockerManager.VolumeCreate(ctx, func(opts *client.VolumeCreateOptions) {
		opts.Driver = string(driver)
		opts.DriverOpts = driverOpts
		opts.Labels = map[string]string{
			docker.StackLabelNamespace: project.Key,
		}
		opts.Name = project.Key + "_default"
	})
	if err != nil {
		return nil, nil, apperrors.Wrap(err)
	}

	timeNow := timeutil.NowUTC()
	setting := &entity.Setting{
		ID:              gofn.Must(ulid.NewStringULID()),
		Scope:           base.ObjectScopeProject,
		ObjectID:        project.ID,
		Type:            base.SettingTypeClusterVolume,
		Kind:            string(driver),
		Status:          base.SettingStatusActive,
		Name:            "default",
		Inheritable:     true,
		Default:         true,
		Version:         entity.CurrentClusterVolumeVersion,
		UpdateVer:       1,
		CreatedAt:       timeNow,
		UpdatedAt:       timeNow,
		CurrentObjectID: project.ID,
	}

	clusterVolume := &entity.ClusterVolume{
		RefID: dockerhelper.GetVolumeID(&createResp.Volume),
	}
	setting.MustSetData(clusterVolume)

	return setting, createResp, nil
}

func (s *service) ListProjectVolumes(
	ctx context.Context,
	db database.IDB,
	project *entity.Project,
	extraOpts ...bunex.SelectQueryOption,
) (settings []*entity.Setting, volumes map[string]*volume.Volume, err error) {
	settings, _, err = s.settingRepo.List(ctx, db, project.GetObjectScope(), nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeClusterVolume),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
	)
	if err != nil {
		return nil, nil, apperrors.Wrap(err)
	}
	if len(settings) == 0 {
		return nil, nil, nil
	}

	volIDs := make([]string, 0, len(settings))
	for _, setting := range settings {
		volIDs = append(volIDs, setting.MustAsClusterVolume().RefID)
	}

	volList, err := s.dockerManager.VolumeListByIDs(ctx, volIDs)
	if err != nil {
		return nil, nil, apperrors.Wrap(err)
	}

	volumes = make(map[string]*volume.Volume, len(settings))
	for _, volID := range volIDs {
		vol, found := gofn.FindPtr(volList.Items, func(vol *volume.Volume) bool {
			return dockerhelper.GetVolumeID(vol) == volID
		})
		if found {
			volumes[volID] = &vol
		}
	}

	return settings, volumes, nil
}

func (s *service) RemoveAllProjectVolumes(
	ctx context.Context,
	db database.IDB,
	project *entity.Project,
	force bool,
) error {
	settings, volumes, err := s.ListProjectVolumes(ctx, db, project)
	if err != nil {
		return apperrors.Wrap(err)
	}

	for _, setting := range settings {
		if setting.ObjectID != project.ID { // imported/inherited volume, skip it
			continue
		}
		vol := volumes[setting.MustAsClusterVolume().RefID]
		if vol == nil {
			continue
		}

		// TODO: if the vol is a local and uses a custom directory, we may need to
		// remove the directory manually.

		_, e := s.dockerManager.VolumeRemove(ctx, dockerhelper.GetVolumeID(vol), force)
		if e != nil && !errors.Is(e, apperrors.ErrNotFound) {
			err = errors.Join(err, e)
		}
	}
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}
