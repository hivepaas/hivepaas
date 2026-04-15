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
	storageSettingName = "Storage settings"
)

func (s *service) initDefaultStorageSettings(
	ctx context.Context,
	db database.IDB,
	timeNow time.Time,
) (err error) {
	storageSetting := &entity.Setting{
		ID:              gofn.Must(ulid.NewStringULID()),
		Scope:           base.SettingScopeGlobal,
		Type:            base.SettingTypeStorageSettings,
		Status:          base.SettingStatusActive,
		Name:            storageSettingName,
		AvailInProjects: true,
		Default:         true,
		Version:         entity.CurrentStorageSettingsVersion,
		CreatedAt:       timeNow,
		UpdatedAt:       timeNow,
	}
	storage := &entity.StorageSettings{}

	storageCfg := &config.Current.Storage
	if storageCfg.BindSource != "" {
		storage.BindSettings = &entity.StorageBindSettings{
			BaseDirs:            []string{storageCfg.BindSource},
			AppsMustUseSubPaths: true,
		}
	}
	if storageCfg.Volume != "" {
		storage.VolumeSettings = &entity.StorageVolumeSettings{
			Volumes:             []string{storageCfg.Volume},
			AppsMustUseSubPaths: true,
		}
	}

	storageSetting.MustSetData(storage)

	err = s.settingRepo.Insert(ctx, db, storageSetting)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}
