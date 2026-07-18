package entity

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

const (
	CurrentSystemBackupVersion = 1
)

var _ = registerSettingParser(base.SettingTypeSystemBackup, &systemBackupParser{})

type systemBackupParser struct {
}

func (s *systemBackupParser) New() SettingData {
	return &SystemBackup{}
}

type SystemBackup struct {
	Schedule       SchedJobSchedule         `json:"schedule"`
	Compression    SystemBackupCompression  `json:"compression,omitempty"`
	Encryption     SystemBackupEncryption   `json:"encryption,omitempty"`
	CloudStorage   SystemBackupCloudStorage `json:"cloudStorage,omitempty"`
	DBBackupConfig SystemBackupDBConfig     `json:"dbBackupConfig"`
	Notification   *BaseEventNotification   `json:"notification,omitempty"`
}

type SystemBackupCompression struct {
	Format base.FileCompressionFormat `json:"format,omitempty"`
}

type SystemBackupEncryption struct {
	Format base.FileEncryptionFormat `json:"format,omitempty"`
	Secret EncryptedField            `json:"secret,omitzero"`
}

type SystemBackupCloudStorage struct {
	ID             string `json:"id,omitempty"` // can be S3 setting ID
	Bucket         string `json:"bucket,omitempty"`
	DestinationDir string `json:"destinationDir,omitempty"`
}

type SystemBackupDBConfig struct {
	BackupDeletedObjects bool `json:"backupDeletedObjects"`
}

func (s *SystemBackup) GetType() base.SettingType {
	return base.SettingTypeSystemBackup
}

func (s *SystemBackup) GetRefObjectIDs() *RefObjectIDs {
	refIDs := &RefObjectIDs{}
	if s.CloudStorage.ID != "" {
		refIDs.RefSettingIDs = append(refIDs.RefSettingIDs, s.CloudStorage.ID)
	}
	if s.Notification != nil {
		refIDs.AddRefIDs(s.Notification.GetRefObjectIDs())
	}
	return refIDs
}

func (s *SystemBackup) GetResourceLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().GetResourceLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *SystemBackup) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentSystemBackupVersion {
		return false, nil
	}
	if setting.Version > CurrentSystemBackupVersion {
		return false, apperrors.Wrap(apperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentSystemBackupVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}

func (s *SystemBackup) Decrypt() error {
	_, err := s.Encryption.Secret.GetPlain()
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (s *Setting) AsSystemBackup() (*SystemBackup, error) {
	return parseSettingAs[*SystemBackup](s)
}

func (s *Setting) MustAsSystemBackup() *SystemBackup {
	return gofn.Must(s.AsSystemBackup())
}
