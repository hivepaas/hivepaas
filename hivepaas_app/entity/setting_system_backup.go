package entity

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
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

// SystemBackup is what the system backup takes - HivePaaS's database, the
// configuration spec of the whole installation, or both - and the backup
// repository it goes into, one snapshot a run.
type SystemBackup struct {
	Schedule    SchedJobSchedule `json:"schedule"`
	IncludeDB   bool             `json:"includeDB,omitempty"`
	IncludeSpec bool             `json:"includeSpec,omitempty"`
	// SpecSecrets is how the spec holds secrets, a spec export's secrets mode:
	// encrypted, omit or plaintext. SpecPassphrase is what encrypted uses.
	SpecSecrets    string         `json:"specSecrets,omitempty"`
	SpecPassphrase EncryptedField `json:"specPassphrase,omitzero"`
	// TargetRepository is a backup repository at the global scope.
	TargetRepository ObjectID               `json:"targetRepository,omitzero"`
	Notification     *BaseEventNotification `json:"notification,omitempty"`
}

// What a system backup's snapshot holds, by name.
const (
	SystemBackupIncludeDB   = "database"
	SystemBackupIncludeSpec = "spec"
)

// Includes is what a run takes, by name.
func (s *SystemBackup) Includes() []string {
	var includes []string
	if s.IncludeDB {
		includes = append(includes, SystemBackupIncludeDB)
	}
	if s.IncludeSpec {
		includes = append(includes, SystemBackupIncludeSpec)
	}
	return includes
}

func (s *SystemBackup) GetType() base.SettingType {
	return base.SettingTypeSystemBackup
}

func (s *SystemBackup) GetRefObjectIDs() *RefObjectIDs {
	refIDs := &RefObjectIDs{}
	if s.TargetRepository.ID != "" {
		refIDs.RefSettingIDs = append(refIDs.RefSettingIDs, s.TargetRepository.ID)
	}
	if s.Notification != nil {
		refIDs.AddRefIDs(s.Notification.GetRefObjectIDs())
	}
	return refIDs
}

func (s *SystemBackup) GetResourceLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().GetResourceLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *SystemBackup) Decrypt() error {
	_, err := s.SpecPassphrase.GetPlain()
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (s *Setting) AsSystemBackup() (*SystemBackup, error) {
	return parseSettingAs[*SystemBackup](s)
}

func (s *Setting) MustAsSystemBackup() *SystemBackup {
	return gofn.Must(s.AsSystemBackup())
}
