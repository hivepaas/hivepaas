// Package walgkit is what the managed engines that archive with WAL-G share:
// engine API 1's contract with the template's scripts - where they are, the
// archive's variables and keys, the jobs that run them and the outputs they
// write - and what the engine makes of it: the app's setting and jobs, its
// status, retention, and changes of its settings. An engine gives its Profile,
// what differs between them; what is one engine's alone stays in its package.
package walgkit

import (
	"regexp"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
)

// Profile is an engine, as what it shares with the others needs it.
type Profile struct {
	Engine templatemodel.ManagedEngine
	// API is the engine API the engine's scripts speak.
	API int
	// Scripts are the scripts the engine API runs, which the template ships as
	// config files in ScriptDir.
	Scripts []string
	// FileOwner is the user and group id the scripts run as, which the secret
	// files are readable by alone.
	FileOwner string
	// BackupPattern is a base backup's name in WAL-G's list, before it reaches
	// WAL-G's arguments.
	BackupPattern *regexp.Regexp
	// DeltaMarker is in a delta backup's name, and in no full backup's; empty
	// for an engine that takes full backups only.
	DeltaMarker string
	// PendingBytesOutput is the walg-health output that says what waits on the
	// node to be archived.
	PendingBytesOutput string
	// Dump is the engine's logical dump.
	Dump DumpCommands
	// RestoreTarget is a restore's target as the variables walg-entrypoint
	// recovers to: which base backup it fetches, and where it stops.
	RestoreTarget func(source *managedmodel.RestoreSource) ([][2]string, error)
	// Members are the database's members as its last health check saw them;
	// nil for an engine without replicas.
	Members func(members []*entity.ManagedMember, health *managedmodel.JobRun) []*managedmodel.MemberStatus
	// Promotion is how a member is promoted; empty, by PromoteScript.
	Promotion Promotion
}

// Promotion is the command that promotes the member it runs in, and the script
// whose mount says the template's scripts can.
type Promotion struct {
	Command  string
	Requires string
}

// DumpCommands are an engine's logical dump: the command that writes it to
// stdout in the database's container, the one that loads it from stdin, and
// the extension of the file it is kept as.
type DumpCommands struct {
	Source  string
	Restore string
	FileExt string
}

const (
	// ScriptDir is where the template's scripts are, beside the artifacts.
	ScriptDir = "/opt/hivepaas/bin"
	// SecretsDir is where the credentials are, one file each, readable by the
	// scripts' user only. walg-run exports them for the one process it runs.
	SecretsDir = "/run/secrets/walg" //nolint:gosec // G101: a directory
	// VolumeArchiveDir is where a volume repository's directory is mounted.
	VolumeArchiveDir = "/var/lib/hivepaas-walg"

	SecretAccessKey = "WALG_AWS_ACCESS_KEY_ID"     //nolint:gosec // G101: a secret's name
	SecretSecretKey = "WALG_AWS_SECRET_ACCESS_KEY" //nolint:gosec // G101: a secret's name
	SecretSodiumKey = "WALG_LIBSODIUM_KEY"         //nolint:gosec // G101: a secret's name
	// EnvSecretsDir tells walg-run where the credentials are.
	EnvSecretsDir = "HIVEPAAS_WALG_SECRETS_DIR" //nolint:gosec // G101: a variable's name

	DefaultArchiveTimeout      = time.Minute
	DefaultRetainFull          = 7
	DefaultBaseBackupSchedule  = "0 2 * * *"
	DefaultRestoreTestSchedule = "0 4 * * 0"
	// DefaultDumpSchedule runs after the base backup's default, 0 2 * * *.
	DefaultDumpSchedule = "0 3 * * *"
	HealthInterval      = 5 * time.Minute

	// EnvRetainFull and EnvRestoreCheckQuery are what walg-nightly and
	// walg-restore-test read their options from, in their job's environment.
	EnvRetainFull        = "WALG_RETAIN_FULL"
	EnvRestoreCheckQuery = "HIVEPAAS_RESTORE_CHECK_QUERY"
	// EnvRestoreSourceHost and EnvRestoreScratchDir tell walg-restore-test it
	// runs in a verifier: the live database it makes its restore point on is at
	// the host given - its app's key, on the project's network - and it restores
	// into the directory given, not beside the data.
	EnvRestoreSourceHost = "HIVEPAAS_RESTORE_SOURCE_HOST"
	EnvRestoreScratchDir = "HIVEPAAS_RESTORE_SCRATCH_DIR"
	// VerifierScratchDir is where a verifier restores into, on a volume of its own.
	VerifierScratchDir = "/var/lib/hivepaas-verify"

	NightlyTimeout     = 6 * time.Hour
	HealthTimeout      = 2 * time.Minute
	RestoreTestTimeout = 6 * time.Hour
)

// The outputs engine API 1's scripts write, as KEY=value lines: part of the
// contract between the template's scripts and the engine.
const (
	// walg-health; what waits to be archived is the Profile's.
	OutLastArchivedAt = "LAST_ARCHIVED_AT"
	OutFailedCount    = "FAILED_COUNT"
	OutDBBytes        = "DB_BYTES"
	OutFreeBytes      = "FREE_BYTES"
	// walg-nightly
	OutBackupName     = "BACKUP_NAME"
	OutBackupAt       = "BACKUP_AT"
	OutBackupBytes    = "BACKUP_BYTES"
	OutOldestBackupAt = "OLDEST_BACKUP_AT"
	// OutBackups is every base backup kept, oldest first: space-separated
	// name@time@bytes, the time RFC 3339. Both of walg-nightly's steps print it,
	// the retention's after it deleted.
	OutBackups = "BACKUPS"
	// walg-restore-test
	OutRestoredTo      = "RESTORED_TO"
	OutDurationSeconds = "DURATION_SECONDS"
)
