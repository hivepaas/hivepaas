package appmanageduc

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
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
		Postgres: &entity.ManagedPostgres{
			Storage:        entity.ManagedStorage{BackupRepo: entity.ObjectID{ID: "repo-1"}, Prefix: "kopia/walg/db"},
			ArchiveTimeout: timeutil.Duration(time.Minute), RetainFull: 7,
			RestoreTest: entity.ManagedRestoreTest{Mode: entity.ManagedRestoreTestContainer},
			Jobs: entity.ManagedPostgresJobs{Nightly: entity.ObjectID{ID: "job-1"},
				Health: entity.ObjectID{ID: "job-2"}, RestoreTest: entity.ObjectID{ID: "job-3"}},
		},
	})
	job := func(id, cron string) *entity.Setting {
		return setting(id, base.SettingTypeSchedJob, &entity.SchedJob{JobType: base.SchedJobTypeContainerCommand,
			Schedule: &entity.SchedJobSchedule{CronExpr: cron}, ManagedBy: entity.ObjectID{ID: "managed-1"},
			Command: &entity.CommandTemplate{Command: "/opt/hivepaas/bin/walg"}})
	}
	loaded := &loadedManaged{setting: managed, engine: postgres.New(),
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
}

func TestReconfiguringRewritesTheSettingAndItsJobs(t *testing.T) {
	loaded := loadedPostgres(t)

	changed, rescheduled, err := reconfigureManaged(loaded, &appmanageddto.UpdateManagedSettingsReq{
		UpdateVer: 3, RetainFull: new(14), BaseBackupSchedule: new("30 1 * * *")}, nil)

	assert.NoError(t, err)
	assert.Equal(t, 4, loaded.setting.UpdateVer)
	assert.Equal(t, 14, loaded.setting.MustAsAppManagedSettings().Postgres.RetainFull, "stored, not only parsed")
	assert.Equal(t, []string{"job-1"}, settingIDs(changed))
	assert.Equal(t, []string{"job-1"}, settingIDs(rescheduled))
	assert.Equal(t, "30 1 * * *", changed[0].MustAsSchedJob().Schedule.CronExpr)
}

func TestReconfiguringAnOlderVersionIsRefused(t *testing.T) {
	_, _, err := reconfigureManaged(loadedPostgres(t), &appmanageddto.UpdateManagedSettingsReq{UpdateVer: 2,
		RetainFull: new(14)}, nil)

	assert.ErrorIs(t, err, hperrors.ErrUpdateVerMismatched)
}

func settingIDs(settings []*entity.Setting) []string {
	out := make([]string, 0, len(settings))
	for _, s := range settings {
		out = append(out, s.ID)
	}
	return out
}
