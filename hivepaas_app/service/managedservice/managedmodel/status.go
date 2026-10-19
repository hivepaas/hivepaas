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
	// Schedules are the engine's jobs' schedules, by the job's id: when retention
	// next runs decides until when a point can be restored.
	Schedules map[string]*entity.SchedJobSchedule
	Now       time.Time
}

const (
	ArchivingOK      = "ok"
	ArchivingFailing = "failing"
	// ArchivingUnknown is a database whose health has not been checked yet.
	ArchivingUnknown = "unknown"

	// SuggestVerifier suggests running the restore test in a verifier app.
	SuggestVerifier = "verifier"
	// SuggestDumpApart says the dump fires in the same minute as the base
	// backup: both read the whole database at once.
	SuggestDumpApart = "dump-apart"
)

// Status is a managed database's state as its engine reads it from its jobs'
// last runs. It is read, never stored, and computed without asking the database.
type Status struct {
	Archiving ArchivingStatus
	// RestoreWindow is what a restore may target: from the oldest base backup
	// kept to the last log archived (WAL for Postgres). Nil before the first base
	// backup.
	RestoreWindow *TimeWindow
	LastBackup    *BackupStatus
	// Backups are the base backups kept, oldest first, as the last nightly run
	// left them: what a restore to a backup may name.
	Backups         []BackupRef
	LastRestoreTest *RestoreTestStatus
	// LastDump is the dump's last run, nil while dumps are off or before the
	// first.
	LastDump *DumpStatus
	// Placement is where the archive is from the database: off-site, shared,
	// same-node.
	Placement   string
	Suggestions []string
	// RestorePoints are the points recorded, oldest first, each with until when
	// it can be restored.
	RestorePoints []*RestorePointStatus
	// HeldBack is what the kept points make retention keep beyond its own,
	// nil when they keep nothing more.
	HeldBack *HeldBack
	// Members are the replicas, each as the primary's last health check saw it.
	Members []*MemberStatus
}

// The states of a member: the primary's pg_stat_replication's, and these.
const (
	// MemberUnknown is a member no health check has looked at yet.
	MemberUnknown = "unknown"
	// MemberDisconnected is a member the primary has no connection from.
	MemberDisconnected = "disconnected"
)

// MemberStatus is a member, and how far behind the primary it was when the
// health check last looked: LagBytes of WAL, and LagSeconds of time, nil while
// nothing is written.
type MemberStatus struct {
	Member     *entity.ManagedMember
	State      string
	LagBytes   int64
	LagSeconds *float64
	CheckedAt  *time.Time
}

// RestorePointStatus is a point, and until when it can be restored: as long as
// the base backup before it is kept. RestorableUntil is nil for a point no
// backup kept reaches, and an estimate from the nightly job's next runs
// otherwise.
type RestorePointStatus struct {
	Point           *entity.ManagedRestorePoint
	Kept            bool
	RestorableUntil *time.Time
}

// HeldBack is the base backups kept for points beyond the retainFull newest,
// and their size. The log after them is held back too, and not counted.
type HeldBack struct {
	Backups int
	Bytes   int64
}

type ArchivingStatus struct {
	State          string
	CheckedAt      *time.Time
	LastArchivedAt *time.Time
	FailedCount    int
	// PendingBytes is what waits on the node to be archived: pg_wal, or binary
	// logs not yet pushed.
	PendingBytes int64
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
	Name  string
	At    time.Time
	Bytes int64
}

// DumpStatus is a dump's run, and the snapshot it took when it succeeded.
type DumpStatus struct {
	Succeeded  bool
	EndedAt    time.Time
	SnapshotID string
	Bytes      int64
}

type RestoreTestStatus struct {
	Passed     bool
	EndedAt    time.Time
	Duration   time.Duration
	RestoredTo string
}
