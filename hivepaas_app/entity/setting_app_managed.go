package entity

import (
	"encoding/json"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

const (
	CurrentAppManagedVersion = 2

	// ManagedRestoreTestContainer runs the restore test in the database's own
	// container; ManagedRestoreTestVerifier in a verifier app, scaled up for it.
	ManagedRestoreTestContainer = "container"
	ManagedRestoreTestVerifier  = "verifier"

	// Where an archive is from the database it keeps: in object storage, on a
	// volume any node mounts, or on a volume of the database's own node.
	ManagedPlacementOffSite  = "off-site"
	ManagedPlacementShared   = "shared"
	ManagedPlacementSameNode = "same-node"

	// What a restore recovers to: everything archived, a time, a restore point,
	// or a base backup as it was taken.
	ManagedRestoreToLatest = "latest"
	ManagedRestoreToTime   = "time"
	ManagedRestoreToPoint  = "point"
	ManagedRestoreToBackup = "backup"
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
	Engine    string       `json:"engine"`
	EngineAPI int          `json:"engineApi"`
	Locks     ManagedLocks `json:"locks"`
	// Database is the engine's state of the database it looks after, whatever
	// the engine; nil for a member, whose state is its primary's.
	Database *ManagedDatabase `json:"database,omitempty"`
	// LegacyPostgres is where version 1 kept Database, read only to move it.
	LegacyPostgres *ManagedDatabase `json:"postgres,omitempty"`
	// SwitchedTo is the database restored from this one that its consumers were
	// switched to: this one's jobs are paused since, and it is left running for
	// the person to stop or delete.
	SwitchedTo *ManagedSwitch `json:"switchedTo,omitempty"`
	// MemberOf is the primary a member of a cluster belongs to, nil for a
	// database of its own. A member's setting holds its locks, and no engine
	// state: that is its primary's.
	MemberOf *ManagedMemberOf `json:"memberOf,omitempty"`
}

// ManagedMemberOf is the primary a member belongs to, and its role there.
type ManagedMemberOf struct {
	App  ObjectID `json:"app"`
	Role string   `json:"role"`
}

// The roles of a member. ManagedRoleReplica is an asynchronous read replica: a
// standby that streams from the cluster's primary, the archive its fallback.
// ManagedRolePrimary is the member promoted to the cluster's primary.
// ManagedRoleFenced is a member that was the primary before another was
// promoted: stopped, and kept from starting its server.
const (
	ManagedRoleReplica = "replica"
	ManagedRolePrimary = "primary"
	ManagedRoleFenced  = "fenced"
)

// ManagedMember is a member of a primary's cluster: an app, a logical child of the
// primary, on a volume of its own.
type ManagedMember struct {
	App ObjectID `json:"app"`
	// Key is the member app's key, which the primary knows its connection by.
	Key  string `json:"key"`
	Role string `json:"role"`
	// Volume is the member's data volume, and Node the node it is pinned to: an
	// id, or a label.
	Volume ObjectID  `json:"volume"`
	Node   string    `json:"node,omitempty"`
	At     time.Time `json:"at"`
	// ServerID is the member's server id, for an engine whose servers each have
	// one (MariaDB's): unique in the cluster, the cluster's app's being 1.
	ServerID int `json:"serverId,omitempty"`
}

// ManagedSwitch is a switch of a database's consumers to the database restored
// from it, and when.
type ManagedSwitch struct {
	App ObjectID  `json:"app"`
	At  time.Time `json:"at"`
}

// ManagedLocks are what the engine owns of the app, recorded when it was made:
// the settings screens refuse to change any of it, knowing nothing of the engine.
type ManagedLocks struct {
	// Image is the image the app was created with, and ImageLine its template
	// version: the repository stays, and the tag moves only within the line and
	// the tag's family.
	Image     string `json:"image"`
	ImageLine string `json:"imageLine"`
	// Command is the engine's command line: the app's starts with it, and what a
	// person adds comes after. The entrypoint stays as the app was made: the
	// engine's, walg-entrypoint (the image's, for a database made before it).
	Command string `json:"command,omitempty"`
	// Mounts are the targets of the mounts the engine relies on, the data's
	// among them: each keeps its source.
	Mounts []string `json:"mounts,omitempty"`
	// EnvKeys are the variables the engine set or relies on, which keep their
	// values.
	EnvKeys []string `json:"envKeys,omitempty"`
	// Settings are the ids of the secrets, config files and setting mounts the app
	// was built with: the engine's credentials and scripts.
	Settings []string `json:"settings,omitempty"`
}

// ManagedDatabase is what the engine keeps of a database: its archive, its
// jobs, its points and its members. None of it is one engine's: what differs
// between engines, such as a point's position, is a value the engine defines.
type ManagedDatabase struct {
	// Storage is where WAL-G writes. It is chosen when the app is created, and
	// its prefix never changes: the archive is there.
	Storage ManagedStorage `json:"storage"`
	// EncryptionKey is WAL-G's libsodium key. Losing it loses the archive.
	EncryptionKey  EncryptedField     `json:"encryptionKey,omitzero"`
	ArchiveTimeout timeutil.Duration  `json:"archiveTimeout"`
	RetainFull     int                `json:"retainFull"`
	RestoreTest    ManagedRestoreTest `json:"restoreTest"`
	// Dump is the logical dump the engine takes beside the archive.
	Dump ManagedDump `json:"dump,omitzero"`
	Jobs ManagedJobs `json:"jobs"`
	// RestorePoints are the moments a person named, to restore to exactly.
	RestorePoints []*ManagedRestorePoint `json:"restorePoints,omitempty"`
	// Restore is where this database was restored from, when it was.
	Restore *ManagedRestore `json:"restore,omitempty"`
	// Members are the database's replicas.
	Members []*ManagedMember `json:"members,omitempty"`
	// BackupFrom is the member the base backups are taken on, to spare the
	// database; empty, they are taken on the database itself.
	BackupFrom ObjectID `json:"backupFrom,omitzero"`
	// Writer is the member promoted to the cluster's primary, which archives and
	// which the engine's commands run on; empty, this app is. This app stays the
	// cluster's: its setting, its jobs, and the address its consumers name.
	Writer ObjectID `json:"writer,omitzero"`
	// FencedAt is when this app was fenced for a promotion: stopped, and kept
	// from starting its server.
	FencedAt time.Time `json:"fencedAt,omitzero"`
	// ReplicationPassword is what the replicas stream with, made with the first
	// one.
	ReplicationPassword EncryptedField `json:"replicationPassword,omitzero"`
}

// WriterKey is the key of the cluster's primary: the writer member's, or app's
// own.
func (s *ManagedDatabase) WriterKey(app *App) string {
	if member := s.Member(s.Writer.ID); member != nil {
		return member.Key
	}
	return app.Key
}

// Member is the member of that app, nil for none.
func (s *ManagedDatabase) Member(appID string) *ManagedMember {
	for _, member := range s.Members {
		if member.App.ID == appID {
			return member
		}
	}
	return nil
}

// ManagedRestorePoint is a moment a person named. A restore targets its
// position, which the engine defines and reads back (a Postgres LSN): the name
// is a label, which Postgres does not keep unique (design §4).
type ManagedRestorePoint struct {
	Name     string    `json:"name"`
	Position string    `json:"position"`
	At       time.Time `json:"at"`
	// LegacyLSN is where version 1 kept Position, read only to move it.
	LegacyLSN string `json:"lsn,omitempty"`
	// By is the user who made it.
	By   string `json:"by,omitempty"`
	Note string `json:"note,omitempty"`
	// KeepUntil holds retention back for the point until then; zero, it is kept
	// as long as retention keeps what it needs.
	KeepUntil time.Time `json:"keepUntil,omitzero"`
	// Origin is what made it: empty for a person, ManagedPointOriginDeploy for
	// a deployment of an app that asks for a point before each of its deploys.
	Origin string `json:"origin,omitempty"`
}

// ManagedPointOriginDeploy is a point a deployment made. The engine removes it
// once no base backup kept reaches it; a person's points are never removed for
// them.
const ManagedPointOriginDeploy = "deploy"

// RestorePoint is the app's point of that name, nil for none.
func (s *ManagedDatabase) RestorePoint(name string) *ManagedRestorePoint {
	for _, point := range s.RestorePoints {
		if point.Name == name {
			return point
		}
	}
	return nil
}

// ManagedRestoreTarget is what a restore recovers to: a kind, and what it needs.
type ManagedRestoreTarget struct {
	Kind string    `json:"kind"`
	Time time.Time `json:"time,omitzero"`
	// Point and Position are a restore point's name and where it is.
	Point    string `json:"point,omitempty"`
	Position string `json:"position,omitempty"`
	Backup   string `json:"backup,omitempty"`
	// LegacyLSN is where version 1 kept Position, read only to move it.
	LegacyLSN string `json:"lsn,omitempty"`
}

// ManagedRestore is where a database restored into a new app came from, and
// when it was ready: recovered, its first base backup taken, and what it was
// restored with gone from it.
type ManagedRestore struct {
	SourceApp ObjectID             `json:"sourceApp"`
	Target    ManagedRestoreTarget `json:"target"`
	ReadyAt   time.Time            `json:"readyAt,omitzero"`
	// SwitchedAt is when the source's consumers were last switched to this one.
	SwitchedAt time.Time `json:"switchedAt,omitzero"`
}

// ManagedStorage is one of: a backup repository reused, S3 or volume, or any
// cloud storage. KeyAuth is a narrower key than the storage's own, when given.
type ManagedStorage struct {
	BackupRepo   ObjectID `json:"backupRepo,omitzero"`
	CloudStorage ObjectID `json:"cloudStorage,omitzero"`
	KeyAuth      ObjectID `json:"keyAuth,omitzero"`
	Prefix       string   `json:"prefix"`
	// Placement is where the archive is from the database, as it was when the
	// app was made: off-site, shared, same-node.
	Placement string `json:"placement,omitempty"`
}

// ManagedDump is a logical dump of the database into a backup repository (the
// engine's dump tool: pg_dump for Postgres), a data backup like any other, run
// on the person's schedule. Without a repository there is no dump: the
// schedule is kept for when there is one.
type ManagedDump struct {
	Repo     ObjectID `json:"repo,omitzero"`
	Schedule string   `json:"schedule,omitempty"`
}

type ManagedRestoreTest struct {
	Mode        string   `json:"mode"`
	VerifierApp ObjectID `json:"verifierApp,omitzero"`
	// NodeID is where the verifier app runs, which should be another node.
	NodeID string `json:"nodeId,omitempty"`
	// KeyAuth is the narrower key the verifier was given, read-only where the
	// storage has one; empty, it has the database's.
	KeyAuth    ObjectID `json:"keyAuth,omitzero"`
	CheckQuery string   `json:"checkQuery,omitempty"`
}

// ManagedJobs are the scheduled jobs the engine made, by role. They are data
// here, not references: each job points at this setting instead.
type ManagedJobs struct {
	Nightly     ObjectID `json:"nightly,omitzero"`
	Health      ObjectID `json:"health,omitzero"`
	RestoreTest ObjectID `json:"restoreTest,omitzero"`
	// Dump is the data backup that dumps the database; empty while dumps are off.
	Dump ObjectID `json:"dump,omitzero"`
}

// UnmarshalJSON reads version 1's names as the current ones: "postgres" as
// "database", a position's "lsn" as "position". Data migration rewrites rows
// only while the stored data version is behind, which not every database's
// is: a row it has not reached is read in the current shape all the same, and
// written in it by SetAppManagedData.
func (s *AppManagedSettings) UnmarshalJSON(data []byte) error {
	type appManagedSettings AppManagedSettings // Without this type's methods, so this one does not call itself.
	if err := json.Unmarshal(data, (*appManagedSettings)(s)); err != nil {
		return hperrors.Wrap(err)
	}
	s.moveLegacyFields()
	return nil
}

// moveLegacyFields moves what version 1 kept under its old names to the
// current ones, and clears the old: written back, the row has the new only.
func (s *AppManagedSettings) moveLegacyFields() {
	if s.Database == nil {
		s.Database = s.LegacyPostgres
	}
	s.LegacyPostgres = nil
	if s.Database == nil {
		return
	}
	for _, point := range s.Database.RestorePoints {
		if point.Position == "" {
			point.Position = point.LegacyLSN
		}
		point.LegacyLSN = ""
	}
	if restore := s.Database.Restore; restore != nil {
		if restore.Target.Position == "" {
			restore.Target.Position = restore.Target.LegacyLSN
		}
		restore.Target.LegacyLSN = ""
	}
}

func (s *AppManagedSettings) GetType() base.SettingType {
	return base.SettingTypeAppManaged
}

func (s *AppManagedSettings) GetRefObjectIDs() *RefObjectIDs {
	refIDs := &RefObjectIDs{}
	if s.SwitchedTo != nil && s.SwitchedTo.App.ID != "" {
		refIDs.RefAppIDs = append(refIDs.RefAppIDs, s.SwitchedTo.App.ID)
	}
	if s.MemberOf != nil && s.MemberOf.App.ID != "" {
		refIDs.RefAppIDs = append(refIDs.RefAppIDs, s.MemberOf.App.ID)
	}
	if s.Database == nil {
		return refIDs
	}
	storage := s.Database.Storage
	for _, id := range []string{storage.BackupRepo.ID, storage.CloudStorage.ID, storage.KeyAuth.ID,
		s.Database.Dump.Repo.ID} {
		if id != "" {
			refIDs.RefSettingIDs = append(refIDs.RefSettingIDs, id)
		}
	}
	if id := s.Database.RestoreTest.VerifierApp.ID; id != "" {
		refIDs.RefAppIDs = append(refIDs.RefAppIDs, id)
	}
	for _, member := range s.Database.Members {
		refIDs.RefAppIDs = append(refIDs.RefAppIDs, member.App.ID)
		if member.Volume.ID != "" {
			refIDs.RefSettingIDs = append(refIDs.RefSettingIDs, member.Volume.ID)
		}
	}
	return refIDs
}

func (s *AppManagedSettings) GetResourceLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().GetResourceLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *AppManagedSettings) Decrypt() error {
	if s.Database == nil {
		return nil
	}
	if _, err := s.Database.EncryptionKey.GetPlain(); err != nil {
		return hperrors.Wrap(err)
	}
	if _, err := s.Database.ReplicationPassword.GetPlain(); err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

// SetAppManagedData stores the engine's state in its setting at the current
// version, which its writers store with the data: a row read in an older
// shape is written in this one.
func (s *Setting) SetAppManagedData(data *AppManagedSettings) error {
	s.Version = CurrentAppManagedVersion
	return s.SetData(data)
}

func (s *Setting) AsAppManagedSettings() (*AppManagedSettings, error) {
	return parseSettingAs[*AppManagedSettings](s)
}

func (s *Setting) MustAsAppManagedSettings() *AppManagedSettings {
	return gofn.Must(s.AsAppManagedSettings())
}
