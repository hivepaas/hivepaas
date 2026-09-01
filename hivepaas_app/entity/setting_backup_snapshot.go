package entity

import (
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

const (
	CurrentBackupSnapshotVersion = 1
)

var _ = registerSettingParser(base.SettingTypeBackupSnapshot, &backupSnapshotParser{})

type backupSnapshotParser struct {
}

func (s *backupSnapshotParser) New() SettingData {
	return &BackupSnapshot{}
}

type BackupSnapshot struct {
	ID          string    `json:"id"`
	ShortID     string    `json:"shortId"`
	Description string    `json:"description,omitempty"`
	Time        time.Time `json:"time"`
	Paths       []string  `json:"paths"`
	Hostname    string    `json:"hostname"`
	SizeBytes   int64     `json:"sizeBytes,omitempty"`
}

func (s *BackupSnapshot) GetType() base.SettingType {
	return base.SettingTypeBackupSnapshot
}

func (s *BackupSnapshot) GetRefObjectIDs() *RefObjectIDs {
	return &RefObjectIDs{}
}

func (s *BackupSnapshot) GetResourceLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().GetResourceLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *Setting) AsBackupSnapshot() (*BackupSnapshot, error) {
	return parseSettingAs[*BackupSnapshot](s)
}

func (s *Setting) MustAsBackupSnapshot() *BackupSnapshot {
	return gofn.Must(s.AsBackupSnapshot())
}
