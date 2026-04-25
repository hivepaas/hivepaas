package entity

import (
	"path/filepath"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/pkg/unit"
)

const (
	CurrentStorageSettingsVersion = 1
)

var _ = registerSettingParser(base.SettingTypeStorageSettings, &storageSettingsParser{})

type storageSettingsParser struct {
}

func (s *storageSettingsParser) New() SettingData {
	return &StorageSettings{}
}

type StorageSettings struct {
	BindSettings          *StorageBindSettings          `json:"bindSettings"`
	VolumeSettings        *StorageVolumeSettings        `json:"volumeSettings"`
	ClusterVolumeSettings *StorageClusterVolumeSettings `json:"clusterVolumeSettings"`
	TmpfsSettings         *StorageTmpfsSettings         `json:"tmpfsSettings"`
}

type StorageBindSettings struct {
	Enabled             bool     `json:"enabled,omitempty"`
	BaseDirs            []string `json:"baseDirs"`
	BaseSubpath         string   `json:"baseSubpath"`
	AppsMustUseSubPaths bool     `json:"appsMustUseSubPaths"`
}

func (s *StorageBindSettings) CaclRequiredSubpath(app *App) string {
	if !s.Enabled || len(s.BaseDirs) == 0 {
		return ""
	}
	if !s.AppsMustUseSubPaths {
		return s.BaseSubpath
	}
	return filepath.Join(s.BaseSubpath, app.Project.Key, app.LocalKey)
}

type StorageVolumeSettings struct {
	Enabled             bool          `json:"enabled,omitempty"`
	Volumes             ObjectIDSlice `json:"volumes"`
	BaseSubpath         string        `json:"baseSubpath"`
	AppsMustUseSubPaths bool          `json:"appsMustUseSubPaths"`
}

func (s *StorageVolumeSettings) CaclRequiredSubpath(app *App) string {
	if !s.Enabled || len(s.Volumes) == 0 {
		return ""
	}
	if !s.AppsMustUseSubPaths {
		return s.BaseSubpath
	}
	return filepath.Join(s.BaseSubpath, app.Project.Key, app.LocalKey)
}

type StorageClusterVolumeSettings struct {
	Enabled             bool          `json:"enabled,omitempty"`
	Volumes             ObjectIDSlice `json:"volumes"`
	BaseSubpath         string        `json:"baseSubpath"`
	AppsMustUseSubPaths bool          `json:"appsMustUseSubPaths"`
}

func (s *StorageClusterVolumeSettings) CaclRequiredSubpath(app *App) string {
	if !s.Enabled || len(s.Volumes) == 0 {
		return ""
	}
	if !s.AppsMustUseSubPaths {
		return s.BaseSubpath
	}
	return filepath.Join(s.BaseSubpath, app.Project.Key, app.LocalKey)
}

type StorageTmpfsSettings struct {
	Enabled bool          `json:"enabled,omitempty"`
	MaxSize unit.DataSize `json:"maxSize"`
}

func (s *StorageSettings) GetType() base.SettingType {
	return base.SettingTypeStorageSettings
}

func (s *StorageSettings) GetRefObjectIDs() *RefObjectIDs {
	refIDs := &RefObjectIDs{}
	return refIDs
}

func (s *StorageSettings) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentStorageSettingsVersion {
		return false, nil
	}
	if setting.Version > CurrentStorageSettingsVersion {
		return false, apperrors.New(apperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentStorageSettingsVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}

func (s *Setting) AsStorageSettings() (*StorageSettings, error) {
	return parseSettingAs[*StorageSettings](s)
}

func (s *Setting) MustAsStorageSettings() *StorageSettings {
	return gofn.Must(s.AsStorageSettings())
}
