package entity

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

const (
	CurrentAppManagedVersion = 1

	// ManagedRestoreTestContainer runs the restore test in the database's own
	// container; ManagedRestoreTestVerifier in a verifier app, scaled up for it.
	ManagedRestoreTestContainer = "container"
	ManagedRestoreTestVerifier  = "verifier"
)

var _ = registerSettingParser(base.SettingTypeAppManaged, &appManagedSettingsParser{})

type appManagedSettingsParser struct {
}

func (s *appManagedSettingsParser) New() SettingData {
	return &AppManagedSettings{}
}

// AppManagedSettings is what the engine that looks after an app keeps of it: the
// engine and the contract it was created under, and the engine's own state.
// The app's artifacts are not here but in app-artifacts, which mounts them
// whatever the engine.
//
// See docs/superpowers/specs/2026-10-17-managed-templates-design.md §3.
type AppManagedSettings struct {
	Engine    string           `json:"engine"`
	EngineAPI int              `json:"engineApi"`
	Postgres  *ManagedPostgres `json:"postgres,omitempty"`
}

type ManagedPostgres struct {
	// Storage is where WAL-G writes. It is chosen when the app is created, and
	// its prefix never changes: the archive is there.
	Storage ManagedStorage `json:"storage"`
	// EncryptionKey is WAL-G's libsodium key. Losing it loses the archive.
	EncryptionKey  EncryptedField      `json:"encryptionKey,omitzero"`
	ArchiveTimeout timeutil.Duration   `json:"archiveTimeout"`
	RetainFull     int                 `json:"retainFull"`
	RestoreTest    ManagedRestoreTest  `json:"restoreTest"`
	Jobs           ManagedPostgresJobs `json:"jobs"`
}

// ManagedStorage is one of: a backup repository reused, S3 or volume, or any
// cloud storage. KeyAuth is a narrower key than the storage's own, when given.
type ManagedStorage struct {
	BackupRepo   ObjectID `json:"backupRepo,omitzero"`
	CloudStorage ObjectID `json:"cloudStorage,omitzero"`
	KeyAuth      ObjectID `json:"keyAuth,omitzero"`
	Prefix       string   `json:"prefix"`
}

type ManagedRestoreTest struct {
	Mode        string   `json:"mode"`
	VerifierApp ObjectID `json:"verifierApp,omitzero"`
	// NodeID is where the verifier app runs, which should be another node.
	NodeID     string `json:"nodeId,omitempty"`
	CheckQuery string `json:"checkQuery,omitempty"`
}

// ManagedPostgresJobs are the scheduled jobs the engine made, by role. They are
// data here, not references: each job points at this setting instead.
type ManagedPostgresJobs struct {
	Nightly     ObjectID `json:"nightly,omitzero"`
	Health      ObjectID `json:"health,omitzero"`
	RestoreTest ObjectID `json:"restoreTest,omitzero"`
}

func (s *AppManagedSettings) GetType() base.SettingType {
	return base.SettingTypeAppManaged
}

func (s *AppManagedSettings) GetRefObjectIDs() *RefObjectIDs {
	refIDs := &RefObjectIDs{}
	if s.Postgres == nil {
		return refIDs
	}
	storage := s.Postgres.Storage
	for _, id := range []string{storage.BackupRepo.ID, storage.CloudStorage.ID, storage.KeyAuth.ID} {
		if id != "" {
			refIDs.RefSettingIDs = append(refIDs.RefSettingIDs, id)
		}
	}
	if id := s.Postgres.RestoreTest.VerifierApp.ID; id != "" {
		refIDs.RefAppIDs = append(refIDs.RefAppIDs, id)
	}
	return refIDs
}

func (s *AppManagedSettings) GetResourceLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().GetResourceLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *AppManagedSettings) Decrypt() error {
	if s.Postgres == nil {
		return nil
	}
	if _, err := s.Postgres.EncryptionKey.GetPlain(); err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (s *Setting) AsAppManagedSettings() (*AppManagedSettings, error) {
	return parseSettingAs[*AppManagedSettings](s)
}

func (s *Setting) MustAsAppManagedSettings() *AppManagedSettings {
	return gofn.Must(s.AsAppManagedSettings())
}
