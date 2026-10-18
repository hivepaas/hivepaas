package postgres

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
	// defaultDumpSchedule runs after the base backup's default, 0 2 * * *.
	defaultDumpSchedule = "0 3 * * *"
	dumpTimeout         = 6 * time.Hour
	dumpJobName         = "Database dump"

	// dumpCommand writes the database to stdout in pg_dump's custom format, as
	// the image's user over the socket: what the data backup streams into its
	// repository. A shell reads the image's variables.
	dumpCommand = `sh -c 'exec pg_dump --format=custom --username="${POSTGRES_USER:-postgres}" ` +
		`--dbname="${POSTGRES_DB:-${POSTGRES_USER:-postgres}}"'`
	// dumpRestoreCommand loads a dump from stdin into the database of the app it
	// runs in - this one, or any app of the postgres image - replacing what the
	// dump has, in one transaction: a restore that fails leaves the database as
	// it was.
	dumpRestoreCommand = `sh -c 'exec pg_restore --clean --if-exists --no-owner --single-transaction ` +
		`--username="${POSTGRES_USER:-postgres}" --dbname="${POSTGRES_DB:-${POSTGRES_USER:-postgres}}"'`

	// The outputs a data backup's run writes, which the last dump is read from.
	outSnapshotID    = "SNAPSHOT_ID"
	outSnapshotBytes = "SNAPSHOT_SIZE_BYTES"

	// dumpCompareSpan is how far ahead the dump's schedule is compared with the
	// base backup's, and maxCompareFires the most runs of either counted: a job
	// every minute has 10080 in a week.
	dumpCompareSpan = 7 * 24 * time.Hour
	maxCompareFires = 20000
)

// newDumpJob is the data backup that dumps the database into the dump's
// repository on its schedule: the job a person could have made, the engine's.
func newDumpJob(app *entity.App, managedID string, dump entity.ManagedDump, restoring bool,
	now time.Time) *entity.Setting {
	return newJob(app.ID, managedID, dumpJobName, &entity.SchedJob{
		JobType:  base.SchedJobTypeDataBackup,
		Schedule: &entity.SchedJobSchedule{CronExpr: gofn.Coalesce(dump.Schedule, defaultDumpSchedule), InitialTime: now},
		Timeout:  timeutil.Duration(dumpTimeout),
		DataBackup: &entity.SchedJobDataBackup{
			Source:           base.SchedJobDataBackupSourceCommand,
			SourceCommand:    &entity.CommandTemplate{Command: dumpCommand},
			SourceFileName:   app.Key + ".dump",
			RestoreCommand:   &entity.CommandTemplate{Command: dumpRestoreCommand},
			TargetRepository: dump.Repo,
		},
	}, restoring, now)
}

// reconfigureDump applies a change of the dump: a repository chosen makes the
// job, one changed moves it, and none removes it. The schedule is kept whether
// dumps are on or not.
func reconfigureDump(req *managedmodel.ReconfigureReq, resp *managedmodel.ReconfigureResp) error {
	pg, change := req.Managed.Postgres, req.Change
	if change.DumpRepo == nil && change.DumpSchedule == nil {
		return nil
	}
	if change.DumpSchedule != nil {
		pg.Dump.Schedule = *change.DumpSchedule
	}
	if change.DumpRepo != nil {
		pg.Dump.Repo = entity.ObjectID{ID: *change.DumpRepo}
	}
	setting := req.Jobs[pg.Jobs.Dump.ID]
	switch {
	case pg.Dump.Repo.ID == "":
		if setting != nil {
			resp.Deleted = append(resp.Deleted, setting)
		}
		pg.Jobs.Dump = entity.ObjectID{}
	case setting == nil:
		// A database made before dumps were offered kept no schedule.
		pg.Dump.Schedule = gofn.Coalesce(pg.Dump.Schedule, defaultDumpSchedule)
		restoring := pg.Restore != nil && pg.Restore.ReadyAt.IsZero()
		created := newDumpJob(req.App, req.ManagedID, pg.Dump, restoring, req.Now)
		pg.Jobs.Dump = entity.ObjectID{ID: created.ID}
		resp.Created = append(resp.Created, created)
	default:
		return moveDump(setting, pg.Dump, resp)
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
		SnapshotID: run.Outputs[outSnapshotID],
		Bytes:      outputInt(run.Outputs, outSnapshotBytes),
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
