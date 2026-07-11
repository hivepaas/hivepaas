package clusterserviceimpl

import (
	"context"
	"time"

	"github.com/moby/moby/api/types/volume"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/dockerhelper"
)

func (s *service) SyncVolumes(
	ctx context.Context,
	db database.IDB,
) ([]volume.Volume, error) {
	// 1. Scan docker to get list of volumes
	volList, err := s.dockerManager.VolumeList(ctx)
	if err != nil {
		return nil, apperrors.New(err)
	}

	currentSettingType := base.SettingTypeClusterVolume
	currentSettingVersion := entity.CurrentClusterVolumeVersion

	// 2. Get list of existing settings from DB
	dbSettings, _, err := s.settingRepo.List(ctx, db, nil, nil,
		bunex.SelectWhere("setting.type = ?", currentSettingType),
	)
	if err != nil {
		return nil, apperrors.New(err)
	}

	existingVols := make(map[string]*entity.Setting, len(dbSettings))
	for _, s := range dbSettings {
		existingVols[dockerhelper.ParseID(s.ID)] = s
	}

	// 3. For each docker volume, if not exists in DB, create new setting
	var updatingSettings []*entity.Setting
	for i := range volList.Items {
		vol := &volList.Items[i]
		volID := vol.Name
		if vol.ClusterVolume != nil {
			volID = vol.ClusterVolume.ID
		}
		setting := existingVols[volID]

		if setting == nil {
			setting = &entity.Setting{
				ID:      dockerhelper.WrapVolumeID(volID),
				Scope:   base.ObjectScopeGlobal,
				Type:    currentSettingType,
				Kind:    vol.Driver,
				Status:  base.SettingStatusActive,
				Name:    vol.Name,
				Version: currentSettingVersion,
			}
			volEntity := &entity.ClusterVolume{}
			if err := setting.SetData(volEntity); err != nil {
				return nil, apperrors.New(err)
			}
			updatingSettings = append(updatingSettings, setting)
			continue
		}

		delete(existingVols, volID)
		hasChanged := false
		if setting.Kind != vol.Driver {
			setting.Kind = vol.Driver
			hasChanged = true
		}
		if setting.Name != vol.Name {
			setting.Name = vol.Name
			hasChanged = true
		}
		if hasChanged {
			updatingSettings = append(updatingSettings, setting)
		}
	}

	// 4. All settings that exist in DB but docker swarm need to remove
	timeNow := time.Now()
	for _, s := range existingVols {
		s.DeletedAt = timeNow
		updatingSettings = append(updatingSettings, s)
	}

	// 5. Upsert the settings
	err = s.settingRepo.UpsertMulti(ctx, db, updatingSettings,
		entity.SettingUpsertingConflictCols, entity.SettingUpsertingUpdateCols)
	if err != nil {
		return nil, apperrors.New(err)
	}

	return volList.Items, nil
}
