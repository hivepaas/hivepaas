package entity

import (
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
	BindSettings   *StorageBindSettings   `json:"bindSettings"`
	VolumeSettings *StorageVolumeSettings `json:"volumeSettings"`
	TmpfsSettings  *StorageTmpfsSettings  `json:"tmpfsSettings"`
}

type StorageBindSettings struct {
	AllowAny            bool     `json:"allowAny,omitempty"`
	BaseDirs            []string `json:"baseDirs"`
	AppsMustUseSubPaths bool     `json:"appsMustUseSubPaths"`
}

type StorageVolumeSettings struct {
	AllowAny            bool     `json:"allowAny,omitempty"`
	Volumes             []string `json:"volumes"`
	AppsMustUseSubPaths bool     `json:"appsMustUseSubPaths"`
}

type StorageTmpfsSettings struct {
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
