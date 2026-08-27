package volumeserviceimpl

import (
	"context"
	"errors"

	"github.com/moby/moby/api/types/volume"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/services/docker/dockerhelper"
)

func (s *service) ListProjectEnvVolumes(
	ctx context.Context,
	db database.IDB,
	projectEnv *entity.ProjectEnv,
	extraOpts ...bunex.SelectQueryOption,
) (settings []*entity.Setting, volumes map[string]*volume.Volume, err error) {
	settings, _, err = s.settingRepo.List(ctx, db, projectEnv.GetObjectScope(), nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeClusterVolume),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
	)
	if err != nil {
		return nil, nil, hperrors.Wrap(err)
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
		return nil, nil, hperrors.Wrap(err)
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

func (s *service) RemoveAllProjectEnvVolumes(
	ctx context.Context,
	db database.IDB,
	projectEnv *entity.ProjectEnv,
	force bool,
) error {
	settings, volumes, err := s.ListProjectEnvVolumes(ctx, db, projectEnv)
	if err != nil {
		return hperrors.Wrap(err)
	}

	for _, setting := range settings {
		if setting.ObjectID != projectEnv.ID { // imported/inherited volume, skip it
			continue
		}
		vol := volumes[setting.MustAsClusterVolume().RefID]
		if vol == nil {
			continue
		}

		// TODO: if the vol is a local and uses a custom directory, we may need to
		// remove the directory manually.

		_, e := s.dockerManager.VolumeRemove(ctx, dockerhelper.GetVolumeID(vol), force)
		if e != nil && !errors.Is(e, hperrors.ErrNotFound) {
			err = errors.Join(err, e)
		}
	}
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
