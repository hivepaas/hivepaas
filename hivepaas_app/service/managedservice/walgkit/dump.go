package walgkit

import (
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
)

const (
	dumpTimeout = 6 * time.Hour
	dumpJobName = "Database dump"

	// The outputs a data backup's run writes, which the last dump is read from.
	OutSnapshotID    = "SNAPSHOT_ID"
	OutSnapshotBytes = "SNAPSHOT_SIZE_BYTES"

	// dumpCompareSpan is how far ahead the dump's schedule is compared with the
	// base backup's, and maxCompareFires the most runs of either counted: a job
	// every minute has 10080 in a week.
	dumpCompareSpan = 7 * 24 * time.Hour
	maxCompareFires = 20000
)

// NewDumpJob is the data backup that dumps the database into the dump's
// repository on its schedule: the job a person could have made, the engine's.
func NewDumpJob(p *Profile, app *entity.App, managedID string, dump entity.ManagedDump, restoring bool,
	now time.Time) *entity.Setting {
	return NewJob(app.ID, managedID, dumpJobName, &entity.SchedJob{
		JobType:  base.SchedJobTypeDataBackup,
		Schedule: &entity.SchedJobSchedule{CronExpr: gofn.Coalesce(dump.Schedule, DefaultDumpSchedule), InitialTime: now},
		Timeout:  timeutil.Duration(dumpTimeout),
		DataBackup: &entity.SchedJobDataBackup{
			Source:           base.SchedJobDataBackupSourceCommand,
			SourceCommand:    &entity.CommandTemplate{Command: p.Dump.Source},
			SourceFileName:   app.Key + p.Dump.FileExt,
			RestoreCommand:   &entity.CommandTemplate{Command: p.Dump.Restore},
			TargetRepository: dump.Repo,
		},
	}, restoring, now)
}

// reconfigureDump applies a change of the dump: a repository chosen makes the
// job, one changed moves it, and none removes it. The schedule is kept whether
// dumps are on or not.
func reconfigureDump(p *Profile, req *managedmodel.ReconfigureReq, resp *managedmodel.ReconfigureResp) error {
	db, change := req.Managed.Database, req.Change
	if change.DumpRepo == nil && change.DumpSchedule == nil {
		return nil
	}
	if change.DumpSchedule != nil {
		db.Dump.Schedule = *change.DumpSchedule
	}
	if change.DumpRepo != nil {
		db.Dump.Repo = entity.ObjectID{ID: *change.DumpRepo}
	}
	setting := req.Jobs[db.Jobs.Dump.ID]
	switch {
	case db.Dump.Repo.ID == "":
		if setting != nil {
			resp.Deleted = append(resp.Deleted, setting)
		}
		db.Jobs.Dump = entity.ObjectID{}
	case setting == nil:
		// A database made before dumps were offered kept no schedule.
		db.Dump.Schedule = gofn.Coalesce(db.Dump.Schedule, DefaultDumpSchedule)
		restoring := db.Restore != nil && db.Restore.ReadyAt.IsZero()
		created := NewDumpJob(p, req.App, req.ManagedID, db.Dump, restoring, req.Now)
		db.Jobs.Dump = entity.ObjectID{ID: created.ID}
		resp.Created = append(resp.Created, created)
	default:
		return moveDump(setting, db.Dump, resp)
	}
	return nil
}

// moveDump gives the dump's job the dump's repository and schedule.
func moveDump(setting *entity.Setting, dump entity.ManagedDump, resp *managedmodel.ReconfigureResp) error {
	job, err := setting.AsSchedJob()
	if err != nil {
		return hperrors.Wrap(err)
	}
	changed, rescheduled := false, false
	if dump.Schedule != "" && job.Schedule != nil && job.Schedule.CronExpr != dump.Schedule {
		job.Schedule.CronExpr = dump.Schedule
		changed, rescheduled = true, true
	}
	if job.DataBackup != nil && job.DataBackup.TargetRepository.ID != dump.Repo.ID {
		job.DataBackup.TargetRepository = dump.Repo
		changed = true
	}
	if !changed {
		return nil
	}
	if err = setting.SetData(job); err != nil {
		return hperrors.Wrap(err)
	}
	setting.UpdateVer++
	resp.Changed = append(resp.Changed, setting)
	if rescheduled {
		resp.Rescheduled = append(resp.Rescheduled, setting)
	}
	return nil
}

// dumpStatus is the dump's last run, and the snapshot it took.
func dumpStatus(run *managedmodel.JobRun) *managedmodel.DumpStatus {
	if run == nil {
		return nil
	}
	return &managedmodel.DumpStatus{
		Succeeded:  run.Status == base.TaskStatusDone,
		EndedAt:    run.EndedAt,
		SnapshotID: run.Outputs[OutSnapshotID],
		Bytes:      OutputInt(run.Outputs, OutSnapshotBytes),
	}
}

// firesTogether says whether two schedules fire in the same minute within the
// next dumpCompareSpan.
func firesTogether(a, b *entity.SchedJobSchedule, now time.Time) bool {
	if a == nil || b == nil {
		return false
	}
	first, err := a.ParseCronExpr()
	if err != nil {
		return false
	}
	second, err := b.ParseCronExpr()
	if err != nil {
		return false
	}
	end := now.Add(dumpCompareSpan)
	fires := map[time.Time]bool{}
	for at, n := first.Next(now), 0; !at.IsZero() && at.Before(end) && n < maxCompareFires; at, n = first.Next(at), n+1 {
		fires[at.Truncate(time.Minute)] = true
	}
	for at, n := second.Next(now), 0; !at.IsZero() && at.Before(end) && n < maxCompareFires; at, n = second.Next(at), n+1 {
		if fires[at.Truncate(time.Minute)] {
			return true
		}
	}
	return false
}
