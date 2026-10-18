package managedmodel

import (
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

// JobRun is the last run of one of the engine's jobs that ended: how it ended,
// and the outputs it wrote.
type JobRun struct {
	Status  base.TaskStatus
	EndedAt time.Time
	Outputs map[string]string
}

// StatusReq is the engine's setting and the last run of each of its jobs, by the
// job's id; a job that has not run yet has none.
type StatusReq struct {
	Managed *entity.AppManagedSettings
	Runs    map[string]*JobRun
	Now     time.Time
}

const (
	ArchivingOK      = "ok"
	ArchivingFailing = "failing"
	// ArchivingUnknown is a database whose health has not been checked yet.
	ArchivingUnknown = "unknown"

	// SuggestVerifier suggests running the restore test in a verifier app.
	SuggestVerifier = "verifier"
)

// Status is a managed database's state as its engine reads it from its jobs'
// last runs. It is read, never stored, and computed without asking the database.
type Status struct {
	Archiving ArchivingStatus
	// RestoreWindow is what a restore may target: from the oldest base backup
	// kept to the last archived WAL. Nil before the first base backup.
	RestoreWindow *TimeWindow
	LastBackup    *BackupStatus
	// Backups are the base backups kept, oldest first, as the last nightly run
	// left them: what a restore to a backup may name.
	Backups         []BackupRef
	LastRestoreTest *RestoreTestStatus
	// Placement is where the archive is from the database: off-site, shared,
	// same-node.
	Placement   string
	Suggestions []string
}

type ArchivingStatus struct {
	State          string
	CheckedAt      *time.Time
	LastArchivedAt *time.Time
	FailedCount    int
	PgWalBytes     int64
	// Stale says the health check has not run for a while: what it last said is
	// old news.
	Stale bool
}

type TimeWindow struct {
	From time.Time
	To   time.Time
}

type BackupStatus struct {
	Succeeded bool
	EndedAt   time.Time
	Name      string
	At        time.Time
	Bytes     int64
}

// BackupRef is a base backup by name, and when it was taken.
type BackupRef struct {
	Name string
	At   time.Time
}

type RestoreTestStatus struct {
	Passed     bool
	EndedAt    time.Time
	Duration   time.Duration
	RestoredTo string
}
