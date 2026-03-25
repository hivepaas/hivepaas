package entity

import (
	"time"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
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
	ScheduleInterval      timeutil.Duration      `json:"scheduleInterval"`
	ScheduleFrom          time.Time              `json:"scheduleFrom"`
	DBBackupConfig        *DBBackupConfig        `json:"dbBackupConfig"`
	Compression           bool                   `json:"compression,omitempty"`
	EncryptionSecret      EncryptedField         `json:"encryptionSecret,omitzero"`
	DestinationStorage    ObjectID               `json:"destinationStorage,omitzero"` // can be S3 setting ID
	DestinationStorageDir string                 `json:"destinationStorageDir,omitempty"`
	LocalBackupRetention  timeutil.Duration      `json:"localBackupRetention,omitempty"`
	Notification          *BaseEventNotification `json:"notification,omitempty"`
}

type DBBackupConfig struct {
	BackupDeletedObjects bool `json:"backupDeletedObjects"`
}

func (s *SystemBackup) GetType() base.SettingType {
	return base.SettingTypeSystemBackup
}

func (s *SystemBackup) GetRefObjectIDs() *RefObjectIDs {
	refIDs := &RefObjectIDs{}
	if s.DestinationStorage.ID != "" {
		refIDs.RefSettingIDs = append(refIDs.RefSettingIDs, s.DestinationStorage.ID)
	}
	if s.Notification != nil {
		refIDs.AddRefIDs(s.Notification.GetRefObjectIDs())
	}
	return refIDs
}

func (s *SystemBackup) MustDecrypt() *SystemBackup {
	s.EncryptionSecret.MustGetPlain()
	return s
}

func (s *SystemBackup) Migrate(setting *Setting) (hasChange bool, err error) {
	if CurrentSystemBackupVersion == setting.Version {
		return false, nil
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentSystemBackupVersion
	setting.MustSetData(s)
	return true, nil
}

func (s *Setting) AsSystemBackup() (*SystemBackup, error) {
	return parseSettingAs[*SystemBackup](s)
}

func (s *Setting) MustAsSystemBackup() *SystemBackup {
	return gofn.Must(s.AsSystemBackup())
}
