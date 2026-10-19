package walgkit

import (
	"strconv"
	"strings"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
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

// Status is a managed app's state, read from its jobs' last runs.
func Status(p *Profile, req *managedmodel.StatusReq) (*managedmodel.Status, error) {
	status := &managedmodel.Status{Archiving: managedmodel.ArchivingStatus{State: managedmodel.ArchivingUnknown}}
	db := req.Managed.Database
	if db == nil {
		return status, nil
	}
	status.Placement = db.Storage.Placement
	run := func(job entity.ObjectID) *managedmodel.JobRun {
		if req.Runs == nil {
			return nil
		}
		return req.Runs[job.ID]
	}
	health, nightly, test := run(db.Jobs.Health), run(db.Jobs.Nightly), run(db.Jobs.RestoreTest)

	if health != nil {
		archiving := &status.Archiving
		archiving.State = managedmodel.ArchivingOK
		if health.Status == base.TaskStatusFailed {
			archiving.State = managedmodel.ArchivingFailing
		}
		archiving.CheckedAt = &health.EndedAt
		archiving.LastArchivedAt = OutputTime(health.Outputs, OutLastArchivedAt)
		archiving.FailedCount = int(OutputInt(health.Outputs, OutFailedCount))
		archiving.PendingBytes = OutputInt(health.Outputs, p.PendingBytesOutput)
		archiving.Stale = req.Now.Sub(health.EndedAt) > staleHealthRuns*HealthInterval
	}
	if nightly != nil {
		status.LastBackup = &managedmodel.BackupStatus{
			Succeeded: nightly.Status == base.TaskStatusDone,
			EndedAt:   nightly.EndedAt,
			Name:      nightly.Outputs[OutBackupName],
			Bytes:     OutputInt(nightly.Outputs, OutBackupBytes),
		}
		if at := OutputTime(nightly.Outputs, OutBackupAt); at != nil {
			status.LastBackup.At = *at
		}
		status.RestoreWindow = restoreWindow(nightly, status.Archiving.LastArchivedAt)
		status.Backups = BackupsOf(p, nightly.Outputs[OutBackups])
		status.HeldBack = HeldBack(p, status.Backups, db.RetainFull, db.RestorePoints, req.Now)
	}
	if test != nil {
		status.LastRestoreTest = &managedmodel.RestoreTestStatus{
			Passed:     test.Status == base.TaskStatusDone,
			EndedAt:    test.EndedAt,
			Duration:   time.Duration(OutputInt(test.Outputs, OutDurationSeconds)) * time.Second,
			RestoredTo: test.Outputs[OutRestoredTo],
		}
	}
	status.LastDump = dumpStatus(run(db.Jobs.Dump))
	if p.Members != nil {
		status.Members = p.Members(db.Members, health)
	}
	status.RestorePoints = PointsStatus(p, db.RestorePoints, status.Backups, db.RetainFull,
		req.Schedules[db.Jobs.Nightly.ID], req.Now)
	if db.RestoreTest.Mode != entity.ManagedRestoreTestVerifier && needsVerifier(health, status.LastRestoreTest) {
		status.Suggestions = append(status.Suggestions, managedmodel.SuggestVerifier)
	}
	if db.Jobs.Dump.ID != "" &&
		firesTogether(req.Schedules[db.Jobs.Nightly.ID], req.Schedules[db.Jobs.Dump.ID], req.Now) {
		status.Suggestions = append(status.Suggestions, managedmodel.SuggestDumpApart)
	}
	return status, nil
}

// restoreWindow runs from the oldest base backup kept to the last log archived,
// or to the last base backup when no health check has said yet.
func restoreWindow(nightly *managedmodel.JobRun, lastArchivedAt *time.Time) *managedmodel.TimeWindow {
	from := OutputTime(nightly.Outputs, OutOldestBackupAt)
	if from == nil {
		return nil
	}
	// The latest of what was archived and the newest base backup: the health
	// check that read the archive may be older than a backup taken since, and a
	// backup is restorable to where it ended.
	to := *from
	for _, at := range []*time.Time{lastArchivedAt, OutputTime(nightly.Outputs, OutBackupAt)} {
		if at != nil && at.After(to) {
			to = *at
		}
	}
	return &managedmodel.TimeWindow{From: *from, To: to}
}

// BackupsOf reads OutBackups, leaving out what does not read: a name that is not
// a backup's could reach retention's arguments.
func BackupsOf(p *Profile, value string) []managedmodel.BackupRef {
	var out []managedmodel.BackupRef
	for _, field := range strings.Fields(value) {
		name, rest, found := strings.Cut(field, "@")
		if !found || !p.BackupPattern.MatchString(name) {
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
		dbBytes, freeBytes := OutputInt(health.Outputs, OutDBBytes), OutputInt(health.Outputs, OutFreeBytes)
		if dbBytes > 0 && freeBytes > 0 && float64(freeBytes) < verifierRoomShare*float64(dbBytes) {
			return true
		}
	}
	return test != nil && test.Duration > verifierTestDuration
}

// OutputTime is an output that is an RFC 3339 time, nil for none.
func OutputTime(outputs map[string]string, key string) *time.Time {
	at, err := time.Parse(time.RFC3339, outputs[key])
	if err != nil {
		return nil
	}
	at = at.UTC()
	return &at
}

// OutputInt is an output that is a number, 0 for none.
func OutputInt(outputs map[string]string, key string) int64 {
	n, _ := strconv.ParseInt(outputs[key], 10, 64)
	return n
}
