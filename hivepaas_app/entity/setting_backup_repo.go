package entity

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/services/backup/backupmodel"
)

const (
	CurrentBackupRepoVersion = 1
)

var _ = registerSettingParser(base.SettingTypeBackupRepo, &backupRepoParser{})

type backupRepoParser struct {
}

func (s *backupRepoParser) New() SettingData {
	return &BackupRepo{}
}

type BackupRepo struct {
	Engine      backupmodel.EngineType `json:"engine"`
	Description string                 `json:"description,omitempty"`
	Password    EncryptedField         `json:"password,omitzero"`

	// Storage
	CloudStorage ObjectID `json:"cloudStorage,omitzero"`
	Volume       ObjectID `json:"volume,omitzero"`
	// StoragePrefix locates the repository inside the storage: a key prefix within the S3 bucket,
	// or a sub-directory within the volume. It lets several repositories share one bucket/volume,
	// and it is what points an imported repository at the data that is already there.
	StoragePrefix string `json:"storagePrefix,omitempty"`

	Compression string                 `json:"compression,omitempty"` // "zstd-fastest"
	PackSize    unit.DataSize          `json:"packSize,omitempty"`    // 16, 32, 64
	Retention   *BackupRetentionPolicy `json:"retention,omitempty"`
}

type BackupRetentionPolicy struct {
	KeepLast    int `json:"keepLast,omitempty"`
	KeepDaily   int `json:"keepDaily,omitempty"`
	KeepWeekly  int `json:"keepWeekly,omitempty"`
	KeepMonthly int `json:"keepMonthly,omitempty"`
}

func (s *BackupRepo) GetType() base.SettingType {
	return base.SettingTypeBackupRepo
}

func (s *BackupRepo) GetRefObjectIDs() *RefObjectIDs {
	refIDs := &RefObjectIDs{}
	if s.CloudStorage.ID != "" {
		refIDs.RefSettingIDs = append(refIDs.RefSettingIDs, s.CloudStorage.ID)
	}
	if s.Volume.ID != "" {
		refIDs.RefSettingIDs = append(refIDs.RefSettingIDs, s.Volume.ID)
	}
	return refIDs
}

func (s *BackupRepo) GetResourceLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().GetResourceLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *BackupRepo) Decrypt() error {
	_, err := s.Password.GetPlain()
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (s *Setting) AsBackupRepo() (*BackupRepo, error) {
	return parseSettingAs[*BackupRepo](s)
}

func (s *Setting) MustAsBackupRepo() *BackupRepo {
	return gofn.Must(s.AsBackupRepo())
}
