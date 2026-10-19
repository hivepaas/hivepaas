package postgres

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/executil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
)

// configured is what Configure made: the engine's setting, parsed, and its jobs
// by id.
func configured(t *testing.T, options *managedmodel.Options) (*entity.Setting, *entity.AppManagedSettings,
	map[string]*entity.Setting) {
	t.Helper()
	req := configureReq(t, deploymentSetting())
	req.Prepared.Options = options
	resp, err := New().Configure(req)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	jobs := map[string]*entity.Setting{}
	for _, setting := range resp.Settings[1:] {
		jobs[setting.ID] = setting
	}
	return resp.Settings[0], resp.Settings[0].MustAsAppManagedSettings(), jobs
}

// A database made with a repository for its dumps has a data backup that runs
// pg_dump into it, on the schedule asked, which a restore pipes into pg_restore.
func TestConfigureMakesTheDumpJob(t *testing.T) {
	setting, managed, jobs := configured(t, &managedmodel.Options{DumpRepo: "repo-1", DumpSchedule: "0 */6 * * *"})

	assert.Len(t, jobs, 4)
	assert.Equal(t, entity.ManagedDump{Repo: entity.ObjectID{ID: "repo-1"}, Schedule: "0 */6 * * *"},
		managed.Database.Dump)
	dumpSetting := jobs[managed.Database.Jobs.Dump.ID]
	if !assert.NotNil(t, dumpSetting) {
		return
	}
	assert.Equal(t, string(base.SchedJobTypeDataBackup), dumpSetting.Kind)
	assert.Equal(t, base.SettingStatusActive, dumpSetting.Status)
	job := dumpSetting.MustAsSchedJob()
	assert.Equal(t, base.SchedJobTypeDataBackup, job.JobType)
	assert.Equal(t, setting.ID, job.ManagedBy.ID)
	assert.Equal(t, "app-1", job.App.ID, "a data backup is its app's")
	assert.Equal(t, "0 */6 * * *", job.Schedule.CronExpr)
	assert.Nil(t, job.Command, "a data backup runs its source's command")
	assert.True(t, job.Notification.FailureUseDefault)

	backup := job.DataBackup
	assert.Equal(t, base.SchedJobDataBackupSourceCommand, backup.Source)
	assert.Equal(t, "repo-1", backup.TargetRepository.ID)
	assert.Equal(t, "shop-db.dump", backup.SourceFileName)
	dump, err := executil.CmdSplit(backup.SourceCommand.Command)
	assert.NoError(t, err)
	assert.Equal(t, []string{"sh", "-c"}, dump[:2], "the image's variables are read by a shell")
	assert.Contains(t, dump[2], "pg_dump --format=custom")
	restore, err := executil.CmdSplit(backup.RestoreCommand.Command)
	assert.NoError(t, err)
	assert.Contains(t, restore[2], "pg_restore --clean --if-exists --no-owner")
}

// Without a repository, no dump is made, and the schedule waits for one.
func TestConfigureMakesNoDumpWithoutARepository(t *testing.T) {
	_, managed, jobs := configured(t, nil)

	assert.Len(t, jobs, 3)
	assert.Empty(t, managed.Database.Jobs.Dump.ID)
	assert.Equal(t, entity.ManagedDump{Schedule: defaultDumpSchedule}, managed.Database.Dump)
}

// A restored database's dump waits for it to recover, as its other jobs do.
func TestARestoredDatabasesDumpWaits(t *testing.T) {
	req := configureReq(t, deploymentSetting())
	req.Prepared.Options = &managedmodel.Options{DumpRepo: "repo-1"}
	req.Prepared.Restore = &managedmodel.RestoreSource{SourceAppID: "source-1",
		Target: entity.ManagedRestoreTarget{Kind: entity.ManagedRestoreToLatest}}

	resp, err := New().Configure(req)

	assert.NoError(t, err)
	managed := resp.Settings[0].MustAsAppManagedSettings()
	for _, setting := range resp.Settings[1:] {
		if setting.ID == managed.Database.Jobs.Dump.ID {
			assert.Equal(t, base.SettingStatusDisabled, setting.Status)
			return
		}
	}
	t.Fatal("no dump job")
}

func changeDump(t *testing.T, setting *entity.Setting, managed *entity.AppManagedSettings,
	jobs map[string]*entity.Setting, change *managedmodel.SettingsChange) *managedmodel.ReconfigureResp {
	t.Helper()
	resp, err := New().Reconfigure(&managedmodel.ReconfigureReq{App: &entity.App{ID: "app-1", Key: "shop-db"},
		ManagedID: setting.ID, Managed: managed, Jobs: jobs, Change: change, Now: time.Now().UTC()})
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	return resp
}

// Choosing a repository turns dumps on: the job is made, on the schedule kept.
func TestReconfigureTurnsDumpsOn(t *testing.T) {
	setting, managed, jobs := configured(t, &managedmodel.Options{DumpSchedule: "15 3 * * *"})

	resp := changeDump(t, setting, managed, jobs, &managedmodel.SettingsChange{DumpRepo: new("repo-2")})

	if !assert.Len(t, resp.Created, 1) {
		return
	}
	created := resp.Created[0]
	assert.Equal(t, created.ID, managed.Database.Jobs.Dump.ID)
	assert.Equal(t, "repo-2", managed.Database.Dump.Repo.ID)
	job := created.MustAsSchedJob()
	assert.Equal(t, setting.ID, job.ManagedBy.ID)
	assert.Equal(t, "15 3 * * *", job.Schedule.CronExpr)
	assert.Equal(t, "repo-2", job.DataBackup.TargetRepository.ID)
	assert.Empty(t, resp.Changed)
	assert.Empty(t, resp.Deleted)
}

// A database made before dumps were offered has no schedule kept: it takes the
// default.
func TestReconfigureTurnsDumpsOnWithTheDefaultSchedule(t *testing.T) {
	setting, managed, jobs := configured(t, nil)
	managed.Database.Dump = entity.ManagedDump{}

	resp := changeDump(t, setting, managed, jobs, &managedmodel.SettingsChange{DumpRepo: new("repo-2")})

	if assert.Len(t, resp.Created, 1) {
		assert.Equal(t, defaultDumpSchedule, resp.Created[0].MustAsSchedJob().Schedule.CronExpr)
		assert.Equal(t, defaultDumpSchedule, managed.Database.Dump.Schedule)
	}
}

func TestReconfigureMovesTheDump(t *testing.T) {
	setting, managed, jobs := configured(t, &managedmodel.Options{DumpRepo: "repo-1"})
	dumpID := managed.Database.Jobs.Dump.ID

	resp := changeDump(t, setting, managed, jobs,
		&managedmodel.SettingsChange{DumpRepo: new("repo-2"), DumpSchedule: new("0 */6 * * *")})

	assert.Equal(t, []string{dumpID}, ids(resp.Changed))
	assert.Equal(t, []string{dumpID}, ids(resp.Rescheduled))
	assert.Empty(t, resp.Created)
	job := jobs[dumpID].MustAsSchedJob()
	assert.Equal(t, "repo-2", job.DataBackup.TargetRepository.ID)
	assert.Equal(t, "0 */6 * * *", job.Schedule.CronExpr)
	assert.Equal(t, entity.ManagedDump{Repo: entity.ObjectID{ID: "repo-2"}, Schedule: "0 */6 * * *"},
		managed.Database.Dump)

	// Another repository alone changes the job, but not when it runs.
	resp = changeDump(t, setting, managed, jobs, &managedmodel.SettingsChange{DumpRepo: new("repo-3")})
	assert.Equal(t, []string{dumpID}, ids(resp.Changed))
	assert.Empty(t, resp.Rescheduled)
}

// Clearing the repository turns dumps off: the job goes, the schedule stays.
func TestReconfigureTurnsDumpsOff(t *testing.T) {
	setting, managed, jobs := configured(t, &managedmodel.Options{DumpRepo: "repo-1", DumpSchedule: "0 5 * * *"})
	dumpID := managed.Database.Jobs.Dump.ID

	resp := changeDump(t, setting, managed, jobs, &managedmodel.SettingsChange{DumpRepo: new("")})

	assert.Equal(t, []string{dumpID}, ids(resp.Deleted))
	assert.Empty(t, resp.Changed)
	assert.Empty(t, managed.Database.Jobs.Dump.ID)
	assert.Equal(t, entity.ManagedDump{Schedule: "0 5 * * *"}, managed.Database.Dump)
}

// The schedule of dumps that are off is kept for when they are on.
func TestReconfigureKeepsTheScheduleOfDumpsOff(t *testing.T) {
	setting, managed, jobs := configured(t, nil)

	resp := changeDump(t, setting, managed, jobs, &managedmodel.SettingsChange{DumpSchedule: new("0 1 * * 0")})

	assert.Empty(t, resp.Created)
	assert.Empty(t, resp.Changed)
	assert.Equal(t, "0 1 * * 0", managed.Database.Dump.Schedule)
}

func TestReconfigureRefusesABadDumpSchedule(t *testing.T) {
	setting, managed, jobs := configured(t, &managedmodel.Options{DumpRepo: "repo-1"})

	for _, schedule := range []string{"at three", ""} {
		_, err := New().Reconfigure(&managedmodel.ReconfigureReq{ManagedID: setting.ID, Managed: managed, Jobs: jobs,
			Change: &managedmodel.SettingsChange{DumpSchedule: new(schedule)}})

		assert.ErrorIs(t, err, hperrors.ErrAppTemplateParamInvalid, "%q", schedule)
	}
}

// A dump turned on while the database is still being restored waits for it.
func TestReconfigureTurnsDumpsOnForARestoringDatabase(t *testing.T) {
	setting, managed, jobs := configured(t, nil)
	managed.Database.Restore = &entity.ManagedRestore{SourceApp: entity.ObjectID{ID: "source-1"}}

	resp := changeDump(t, setting, managed, jobs, &managedmodel.SettingsChange{DumpRepo: new("repo-1")})

	if assert.Len(t, resp.Created, 1) {
		assert.Equal(t, base.SettingStatusDisabled, resp.Created[0].Status)
	}
}

// The last dump is read from its run's outputs, the snapshot it took.
func TestStatusReadsTheLastDump(t *testing.T) {
	managed := statusManaged(entity.ManagedRestoreTestContainer)
	managed.Database.Jobs.Dump = entity.ObjectID{ID: "dump-1"}
	ended := time.Date(2026, 10, 18, 3, 1, 0, 0, time.UTC)
	runs := healthyRuns()
	runs["dump-1"] = &managedmodel.JobRun{Status: base.TaskStatusDone, EndedAt: ended,
		Outputs: map[string]string{outSnapshotID: "k1a2b3", outSnapshotBytes: "52428800"}}

	status, err := New().Status(&managedmodel.StatusReq{Managed: managed, Runs: runs, Now: ended.Add(time.Hour)})

	assert.NoError(t, err)
	assert.Equal(t, &managedmodel.DumpStatus{Succeeded: true, EndedAt: ended, SnapshotID: "k1a2b3",
		Bytes: 52428800}, status.LastDump)

	runs["dump-1"] = &managedmodel.JobRun{Status: base.TaskStatusFailed, EndedAt: ended}
	status, err = New().Status(&managedmodel.StatusReq{Managed: managed, Runs: runs, Now: ended.Add(time.Hour)})
	assert.NoError(t, err)
	assert.Equal(t, &managedmodel.DumpStatus{EndedAt: ended}, status.LastDump)
}

// A dump that fires in the same minute as the base backup is told apart: both
// read the whole database at once.
func TestStatusSuggestsTheDumpRunsApart(t *testing.T) {
	managed := statusManaged(entity.ManagedRestoreTestContainer)
	managed.Database.Jobs.Dump = entity.ObjectID{ID: "dump-1"}
	now := time.Date(2026, 10, 18, 12, 0, 0, 0, time.UTC)
	schedules := func(dump string) map[string]*entity.SchedJobSchedule {
		return map[string]*entity.SchedJobSchedule{
			managed.Database.Jobs.Nightly.ID: {CronExpr: "0 2 * * *"},
			"dump-1":                         {CronExpr: dump},
		}
	}

	for dump, together := range map[string]bool{
		"0 2 * * *":   true,
		"0 */2 * * *": true,
		"0 2 * * 3":   true,
		"0 3 * * *":   false,
		"1 2 * * *":   false,
	} {
		status, err := New().Status(&managedmodel.StatusReq{Managed: managed, Runs: healthyRuns(),
			Schedules: schedules(dump), Now: now})

		assert.NoError(t, err)
		assert.Equal(t, together, contains(status.Suggestions, managedmodel.SuggestDumpApart), dump)
	}
}

func contains(items []string, item string) bool {
	for _, one := range items {
		if one == item {
			return true
		}
	}
	return false
}
