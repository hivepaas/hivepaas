package entity

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

const (
	CurrentBackupRepoCleanupVersion = 1
)

var _ = registerSettingParser(base.SettingTypeBackupRepoCleanup, &backupRepoCleanupParser{})

type backupRepoCleanupParser struct {
}

func (s *backupRepoCleanupParser) New() SettingData {
	return &BackupRepoCleanup{}
}

type BackupRepoCleanup struct {
	Schedule     SchedJobSchedule       `json:"schedule"`
	Notification *BaseEventNotification `json:"notification,omitempty"`
}

func (s *BackupRepoCleanup) GetType() base.SettingType {
	return base.SettingTypeBackupRepoCleanup
}

func (s *BackupRepoCleanup) GetRefObjectIDs() *RefObjectIDs {
	refIDs := &RefObjectIDs{}
	if s.Notification != nil {
		refIDs.AddRefIDs(s.Notification.GetRefObjectIDs())
	}
	return refIDs
}

func (s *BackupRepoCleanup) GetResourceLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().GetResourceLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *Setting) AsBackupRepoCleanup() (*BackupRepoCleanup, error) {
	return parseSettingAs[*BackupRepoCleanup](s)
}

func (s *Setting) MustAsBackupRepoCleanup() *BackupRepoCleanup {
	return gofn.Must(s.AsBackupRepoCleanup())
}
