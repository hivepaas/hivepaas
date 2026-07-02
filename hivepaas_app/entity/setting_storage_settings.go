package entity

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
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
	Enabled         bool     `json:"enabled,omitempty"`
	BaseDirs        []string `json:"baseDirs"`
	SubpathTemplate string   `json:"subpathTemplate"`
}

type StorageVolumeSettings struct {
	Enabled         bool          `json:"enabled,omitempty"`
	Volumes         ObjectIDSlice `json:"volumes"`
	SubpathTemplate string        `json:"subpathTemplate"`
}

type StorageClusterVolumeSettings struct {
	Enabled         bool          `json:"enabled,omitempty"`
	Volumes         ObjectIDSlice `json:"volumes"`
	SubpathTemplate string        `json:"subpathTemplate"`
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

func (s *StorageSettings) CalcResLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().CalcResLinks(base.ResourceTypeSetting, setting.ID)
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
