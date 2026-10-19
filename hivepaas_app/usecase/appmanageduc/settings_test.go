package appmanageduc

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/postgres"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appmanageduc/appmanageddto"
)

// loadedPostgres is a managed app as postgres configured it.
func loadedPostgres(t *testing.T) *loadedManaged {
	t.Helper()
	managed := setting("managed-1", base.SettingTypeAppManaged, &entity.AppManagedSettings{
		Engine: "postgres", EngineAPI: 1,
		Locks: entity.ManagedLocks{Image: "postgres:17.6-trixie", ImageLine: "17"},
		Database: &entity.ManagedDatabase{
			Storage:        entity.ManagedStorage{BackupRepo: entity.ObjectID{ID: "repo-1"}, Prefix: "kopia/walg/db"},
			ArchiveTimeout: timeutil.Duration(time.Minute), RetainFull: 7,
			RestoreTest: entity.ManagedRestoreTest{Mode: entity.ManagedRestoreTestContainer},
			Jobs: entity.ManagedJobs{Nightly: entity.ObjectID{ID: "job-1"},
				Health: entity.ObjectID{ID: "job-2"}, RestoreTest: entity.ObjectID{ID: "job-3"}},
		},
	})
	job := func(id, cron string) *entity.Setting {
		return setting(id, base.SettingTypeSchedJob, &entity.SchedJob{JobType: base.SchedJobTypeContainerCommand,
			Schedule: &entity.SchedJobSchedule{CronExpr: cron}, ManagedBy: entity.ObjectID{ID: "managed-1"},
			Command: &entity.CommandTemplate{Command: "/opt/hivepaas/bin/walg"}})
	}
	loaded := &loadedManaged{app: &entity.App{ID: "app-1", Key: "db"}, setting: managed, engine: postgres.New(),
		jobs: []*entity.Setting{job("job-1", "0 2 * * *"), job("job-2", ""), job("job-3", "0 4 * * 0")}}
	loaded.managed = managed.MustAsAppManagedSettings()
	return loaded
}

func TestTheSettingsSayWhatTheEngineKeepsAndWhen(t *testing.T) {
	resp := appmanageddto.TransformManagedSettings(loadedPostgres(t).view())

	assert.Equal(t, 3, resp.UpdateVer)
	assert.Equal(t, "postgres", resp.Engine)
	assert.Equal(t, "repo-1", resp.Storage.BackupRepo.ID)
	assert.Nil(t, resp.Storage.CloudStorage)
	assert.Equal(t, 7, resp.RetainFull)
	assert.Equal(t, "0 2 * * *", resp.BaseBackupSchedule)
	assert.Equal(t, "0 4 * * 0", resp.RestoreTestSchedule)
	assert.Equal(t, "container", resp.RestoreTest.Mode)
	assert.Equal(t, "postgres:17.6-trixie", resp.Locks.Image)
	assert.Equal(t, "job-1", resp.Jobs.Nightly.ID)
	assert.Equal(t, appmanageddto.ManagedCapabilitiesResp{Replicas: true, VerifierApp: true}, resp.Capabilities,
		"what the engine can do")
}

func TestReconfiguringRewritesTheSettingAndItsJobs(t *testing.T) {
	loaded := loadedPostgres(t)

	resp, err := reconfigureManaged(loaded, &appmanageddto.UpdateManagedSettingsReq{
		UpdateVer: 3, RetainFull: new(14), BaseBackupSchedule: new("30 1 * * *")}, nil, time.Now())

	assert.NoError(t, err)
	assert.Equal(t, 4, loaded.setting.UpdateVer)
	assert.Equal(t, 14, loaded.setting.MustAsAppManagedSettings().Database.RetainFull, "stored, not only parsed")
	assert.Equal(t, []string{"job-1"}, settingIDs(resp.Changed))
	assert.Equal(t, []string{"job-1"}, settingIDs(resp.Rescheduled))
	assert.Equal(t, "30 1 * * *", resp.Changed[0].MustAsSchedJob().Schedule.CronExpr)
	assert.Empty(t, resp.Created)
}

// A repository chosen turns dumps on: the engine makes the job, managed by the
// app's setting, and the settings then name both.
func TestReconfiguringTurnsDumpsOn(t *testing.T) {
	loaded := loadedPostgres(t)

	resp, err := reconfigureManaged(loaded, &appmanageddto.UpdateManagedSettingsReq{UpdateVer: 3,
		DumpRepo: &basedto.ObjectIDReq{ID: "repo-2"}, DumpSchedule: new("0 */6 * * *")}, nil, time.Now())

	assert.NoError(t, err)
	if !assert.Len(t, resp.Created, 1) {
		return
	}
	job := resp.Created[0].MustAsSchedJob()
	assert.Equal(t, "managed-1", job.ManagedBy.ID)
	assert.Equal(t, "app-1", job.App.ID)
	assert.Equal(t, "repo-2", job.DataBackup.TargetRepository.ID)
	loaded.jobs = append(loaded.jobs, resp.Created[0])
	settings := appmanageddto.TransformManagedSettings(loaded.view())
	assert.Equal(t, "repo-2", settings.Dump.Repo.ID)
	assert.Equal(t, "0 */6 * * *", settings.Dump.Schedule)
	assert.Equal(t, resp.Created[0].ID, settings.Jobs.Dump.ID)

	// {"id": ""} turns them off again.
	resp, err = reconfigureManaged(loaded, &appmanageddto.UpdateManagedSettingsReq{UpdateVer: 4,
		DumpRepo: &basedto.ObjectIDReq{}}, nil, time.Now())
	assert.NoError(t, err)
	assert.Len(t, resp.Deleted, 1)
	settings = appmanageddto.TransformManagedSettings(loaded.view())
	assert.Nil(t, settings.Dump.Repo)
	assert.Nil(t, settings.Jobs.Dump)
	assert.Equal(t, "0 */6 * * *", settings.Dump.Schedule, "kept for when they are on")
}

func TestReconfiguringAnOlderVersionIsRefused(t *testing.T) {
	_, err := reconfigureManaged(loadedPostgres(t), &appmanageddto.UpdateManagedSettingsReq{UpdateVer: 2,
		RetainFull: new(14)}, nil, time.Now())

	assert.ErrorIs(t, err, hperrors.ErrUpdateVerMismatched)
}

func settingIDs(settings []*entity.Setting) []string {
	out := make([]string, 0, len(settings))
	for _, s := range settings {
		out = append(out, s.ID)
	}
	return out
}

func TestTheSettingsSayWhereARestoredDatabaseCameFrom(t *testing.T) {
	loaded := loadedPostgres(t)
	assert.Nil(t, appmanageddto.TransformManagedSettings(loaded.view()).Restore, "a database not restored")

	at := time.Date(2026, 10, 18, 14, 32, 0, 0, time.UTC)
	loaded.managed.Database.Restore = &entity.ManagedRestore{SourceApp: entity.ObjectID{ID: "app-src"},
		Target: entity.ManagedRestoreTarget{Kind: entity.ManagedRestoreToTime, Time: at}}
	restore := appmanageddto.TransformManagedSettings(loaded.view()).Restore

	if assert.NotNil(t, restore) {
		assert.Equal(t, "app-src", restore.SourceApp.ID)
		assert.Equal(t, "time", restore.Target.Kind)
		assert.Equal(t, &at, restore.Target.Time)
		assert.Nil(t, restore.ReadyAt, "still restoring")
	}
}

// The base backups move to a replica: the nightly job runs there, and the
// settings say where.
func TestReconfiguringMovesTheBaseBackupsToAReplica(t *testing.T) {
	loaded := loadedPostgres(t)
	loaded.managed.Database.Members = []*entity.ManagedMember{{App: entity.ObjectID{ID: "db-r1"}, Key: "db-r1",
		Role: entity.ManagedRoleReplica}}

	resp, err := reconfigureManaged(loaded, &appmanageddto.UpdateManagedSettingsReq{UpdateVer: 3,
		BackupFrom: &basedto.ObjectIDReq{ID: "db-r1"}}, nil, time.Now())

	assert.NoError(t, err)
	assert.Equal(t, []string{"job-1"}, settingIDs(resp.Changed))
	assert.Equal(t, "db-r1", resp.Changed[0].MustAsSchedJob().App.ID)
	settings := appmanageddto.TransformManagedSettings(loaded.view())
	assert.Equal(t, "db-r1", settings.BackupFrom.ID)

	_, err = reconfigureManaged(loaded, &appmanageddto.UpdateManagedSettingsReq{UpdateVer: 4,
		BackupFrom: &basedto.ObjectIDReq{}}, nil, time.Now())
	assert.NoError(t, err)
	assert.Nil(t, appmanageddto.TransformManagedSettings(loaded.view()).BackupFrom)
}
