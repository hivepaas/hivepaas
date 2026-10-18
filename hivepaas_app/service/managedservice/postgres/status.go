package postgres

import (
	"strconv"
	"strings"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
)

// The outputs engine API 1's scripts write, as KEY=value lines: part of the
// contract between the template's scripts and the engine.
const (
	// walg-health
	outLastArchivedAt = "LAST_ARCHIVED_AT"
	outFailedCount    = "FAILED_COUNT"
	outPgWalBytes     = "PG_WAL_BYTES"
	outDBBytes        = "DB_BYTES"
	outFreeBytes      = "FREE_BYTES"
	// walg-nightly
	outBackupName     = "BACKUP_NAME"
	outBackupAt       = "BACKUP_AT"
	outBackupBytes    = "BACKUP_BYTES"
	outOldestBackupAt = "OLDEST_BACKUP_AT"
	// outBackups is every base backup kept, oldest first: space-separated
	// name@time@bytes, the time RFC 3339. Both of walg-nightly's steps print it,
	// the retention's after it deleted.
	outBackups = "BACKUPS"
	// walg-restore-test
	outRestoredTo      = "RESTORED_TO"
	outDurationSeconds = "DURATION_SECONDS"
)

const (
	// staleHealthRuns is how many health intervals without a run make what the
	// last one said old news.
	staleHealthRuns = 3
	// verifierRoomShare is the free space, against the database's size, under
	// which the container's restore test would not fit: the copy, and room to
	// replay into.
	verifierRoomShare = 1.5
	// verifierTestDuration is how long a restore test may keep the database's
	// container busy before the verifier is suggested.
	verifierTestDuration = 30 * time.Minute
)

func (e *Engine) Status(req *managedmodel.StatusReq) (*managedmodel.Status, error) {
	status := &managedmodel.Status{Archiving: managedmodel.ArchivingStatus{State: managedmodel.ArchivingUnknown}}
	pg := req.Managed.Postgres
	if pg == nil {
		return status, nil
	}
	status.Placement = pg.Storage.Placement
	run := func(job entity.ObjectID) *managedmodel.JobRun {
		if req.Runs == nil {
			return nil
		}
		return req.Runs[job.ID]
	}
	health, nightly, test := run(pg.Jobs.Health), run(pg.Jobs.Nightly), run(pg.Jobs.RestoreTest)

	if health != nil {
		archiving := &status.Archiving
		archiving.State = managedmodel.ArchivingOK
		if health.Status == base.TaskStatusFailed {
			archiving.State = managedmodel.ArchivingFailing
		}
		archiving.CheckedAt = &health.EndedAt
		archiving.LastArchivedAt = outputTime(health.Outputs, outLastArchivedAt)
		archiving.FailedCount = int(outputInt(health.Outputs, outFailedCount))
		archiving.PgWalBytes = outputInt(health.Outputs, outPgWalBytes)
		archiving.Stale = req.Now.Sub(health.EndedAt) > staleHealthRuns*healthInterval
	}
	if nightly != nil {
		status.LastBackup = &managedmodel.BackupStatus{
			Succeeded: nightly.Status == base.TaskStatusDone,
			EndedAt:   nightly.EndedAt,
			Name:      nightly.Outputs[outBackupName],
			Bytes:     outputInt(nightly.Outputs, outBackupBytes),
		}
		if at := outputTime(nightly.Outputs, outBackupAt); at != nil {
			status.LastBackup.At = *at
		}
		status.RestoreWindow = restoreWindow(nightly, status.Archiving.LastArchivedAt)
		status.Backups = backupsOf(nightly.Outputs[outBackups])
		status.HeldBack = heldBack(status.Backups, pg.RetainFull, pg.RestorePoints, req.Now)
	}
	if test != nil {
		status.LastRestoreTest = &managedmodel.RestoreTestStatus{
			Passed:     test.Status == base.TaskStatusDone,
			EndedAt:    test.EndedAt,
			Duration:   time.Duration(outputInt(test.Outputs, outDurationSeconds)) * time.Second,
			RestoredTo: test.Outputs[outRestoredTo],
		}
	}
	status.LastDump = dumpStatus(run(pg.Jobs.Dump))
	status.Members = membersStatus(pg.Members, health)
	status.RestorePoints = pointsStatus(pg.RestorePoints, status.Backups, pg.RetainFull,
		req.Schedules[pg.Jobs.Nightly.ID], req.Now)
	if pg.RestoreTest.Mode != entity.ManagedRestoreTestVerifier && needsVerifier(health, status.LastRestoreTest) {
		status.Suggestions = append(status.Suggestions, managedmodel.SuggestVerifier)
	}
	if pg.Jobs.Dump.ID != "" &&
		firesTogether(req.Schedules[pg.Jobs.Nightly.ID], req.Schedules[pg.Jobs.Dump.ID], req.Now) {
		status.Suggestions = append(status.Suggestions, managedmodel.SuggestDumpApart)
	}
	return status, nil
}

// restoreWindow runs from the oldest base backup kept to the last WAL archived,
// or to the last base backup when no health check has said yet.
func restoreWindow(nightly *managedmodel.JobRun, lastArchivedAt *time.Time) *managedmodel.TimeWindow {
	from := outputTime(nightly.Outputs, outOldestBackupAt)
	if from == nil {
		return nil
	}
	// The latest of what was archived and the newest base backup: the health
	// check that read the archiver may be older than a backup taken since, and
	// a backup is restorable to where it ended, its last WAL archived before
	// backup-push returns.
	to := *from
	for _, at := range []*time.Time{lastArchivedAt, outputTime(nightly.Outputs, outBackupAt)} {
		if at != nil && at.After(to) {
			to = *at
		}
	}
	return &managedmodel.TimeWindow{From: *from, To: to}
}

// backupsOf reads outBackups, leaving out what does not read: a name that is not
// a backup's could reach retention's arguments.
func backupsOf(value string) []managedmodel.BackupRef {
	var out []managedmodel.BackupRef
	for _, field := range strings.Fields(value) {
		name, rest, found := strings.Cut(field, "@")
		if !found || !backupPattern.MatchString(name) {
			continue
		}
		at, size, _ := strings.Cut(rest, "@")
		taken, err := time.Parse(time.RFC3339, at)
		if err != nil {
			continue
		}
		ref := managedmodel.BackupRef{Name: name, At: taken.UTC()}
		ref.Bytes, _ = strconv.ParseInt(size, 10, 64)
		out = append(out, ref)
	}
	return out
}

func needsVerifier(health *managedmodel.JobRun, test *managedmodel.RestoreTestStatus) bool {
	if health != nil {
		dbBytes, freeBytes := outputInt(health.Outputs, outDBBytes), outputInt(health.Outputs, outFreeBytes)
		if dbBytes > 0 && freeBytes > 0 && float64(freeBytes) < verifierRoomShare*float64(dbBytes) {
			return true
		}
	}
	return test != nil && test.Duration > verifierTestDuration
}

func outputTime(outputs map[string]string, key string) *time.Time {
	at, err := time.Parse(time.RFC3339, outputs[key])
	if err != nil {
		return nil
	}
	at = at.UTC()
	return &at
}

func outputInt(outputs map[string]string, key string) int64 {
	n, _ := strconv.ParseInt(outputs[key], 10, 64)
	return n
}
