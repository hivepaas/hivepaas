package settingserviceimpl

import (
	"context"
	"time"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/config"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/ulid"
)

const (
	storageSettingName            = "Storage settings"
	storageSettingSubpathTemplate = "project_data/{{project}}/{{env}}/{{app}}"
)

func (s *service) initDefaultStorageSettings(
	ctx context.Context,
	db database.IDB,
	timeNow time.Time,
) (err error) {
	storageSetting := &entity.Setting{
		ID:              gofn.Must(ulid.NewStringULID()),
		Scope:           base.ObjectScopeGlobal,
		Type:            base.SettingTypeStorageSettings,
		Status:          base.SettingStatusActive,
		Name:            storageSettingName,
		AvailInProjects: true,
		Default:         true,
		Version:         entity.CurrentStorageSettingsVersion,
		CreatedAt:       timeNow,
		UpdatedAt:       timeNow,
	}
	storage := &entity.StorageSettings{
		BindSettings: &entity.StorageBindSettings{
			SubpathTemplate: storageSettingSubpathTemplate,
		},
		VolumeSettings: &entity.StorageVolumeSettings{
			SubpathTemplate: storageSettingSubpathTemplate,
		},
		ClusterVolumeSettings: &entity.StorageClusterVolumeSettings{
			SubpathTemplate: storageSettingSubpathTemplate,
		},
		TmpfsSettings: &entity.StorageTmpfsSettings{},
	}

	storageCfg := &config.Current.Storage
	if storageCfg.BindSource != "" {
		storage.BindSettings.Enabled = true
		storage.BindSettings.BaseDirs = []string{storageCfg.BindSource}
	}
	if storageCfg.Volume != "" {
		storage.VolumeSettings.Enabled = true
		storage.VolumeSettings.Volumes = entity.ObjectIDSlice{{ID: storageCfg.Volume}}

		storage.ClusterVolumeSettings.Enabled = true
		storage.ClusterVolumeSettings.Volumes = entity.ObjectIDSlice{{ID: storageCfg.Volume}}
	}

	storageSetting.MustSetData(storage)

	err = s.settingRepo.Insert(ctx, db, storageSetting)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}
