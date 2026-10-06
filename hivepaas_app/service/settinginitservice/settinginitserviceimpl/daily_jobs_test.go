package settinginitserviceimpl

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
)

// newYork is UTC-5, and UTC-4 in daylight saving, from March to November -
// through the dates these tests take.
func newYork(t *testing.T) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	return loc
}

var daily = timeutil.Duration(timeutil.Day)

// A daily job starts at its time of day there, on the day it is there.
func TestDailyJobStartsAtItsTimeOfDayInTheTimezone(t *testing.T) {
	cleanup := dailyJobOf(base.SettingTypeSystemCleanup)
	backup := dailyJobOf(base.SettingTypeSystemBackup)
	// 02:00 UTC on 6 October is 22:00 on the 5th in New York.
	now := time.Date(2026, 10, 6, 2, 0, 0, 0, time.UTC)

	assert.Equal(t, time.Date(2026, 10, 6, 0, 0, 0, 0, time.UTC), cleanup.startIn(now, time.UTC))
	assert.Equal(t, time.Date(2026, 10, 5, 4, 0, 0, 0, time.UTC),
		cleanup.startIn(now, newYork(t)), "midnight of the 5th there")
	assert.Equal(t, time.Date(2026, 10, 5, 4, 30, 0, 0, time.UTC),
		backup.startIn(now, newYork(t)))
	assert.Zero(t, dailyJobOf(base.SettingTypeSchedJob).settingType, "no daily job")
}

// A schedule is the default only daily, from the job's time of day there.
func TestDailyJobKnowsItsDefault(t *testing.T) {
	cleanup := dailyJobOf(base.SettingTypeSystemCleanup)
	ny := newYork(t)
	utcMidnight := time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)

	assert.True(t, cleanup.isDefaultIn(&entity.SchedJobSchedule{Interval: daily, InitialTime: utcMidnight}, time.UTC))
	assert.False(t, cleanup.isDefaultIn(&entity.SchedJobSchedule{Interval: daily, InitialTime: utcMidnight}, ny))
	assert.True(t, cleanup.isDefaultIn(&entity.SchedJobSchedule{Interval: daily,
		InitialTime: time.Date(2026, 9, 30, 4, 0, 0, 0, time.UTC)}, ny), "midnight there")
	for name, s := range map[string]*entity.SchedJobSchedule{
		"another time":     {Interval: daily, InitialTime: utcMidnight.Add(3 * time.Hour)},
		"another interval": {Interval: timeutil.Duration(12 * time.Hour), InitialTime: utcMidnight},
		"a cron":           {CronExpr: "0 0 * * *", InitialTime: utcMidnight},
		"an end":           {Interval: daily, InitialTime: utcMidnight, EndTime: utcMidnight.AddDate(1, 0, 0)},
	} {
		assert.False(t, cleanup.isDefaultIn(s, time.UTC), name)
	}
}

// settingsOfJobs holds the daily jobs' settings by type and their jobs by the
// setting they run. A job is looked up right after its setting: the fake
// answers the job of the setting last looked up.
type settingsOfJobs struct {
	repository.SettingRepo
	byType  map[base.SettingType]*entity.Setting
	jobs    map[string]*entity.Setting
	last    string
	updated []string
}

func (r *settingsOfJobs) GetSingle(_ context.Context, _ database.IDB, _ *entity.ObjectScope, typ base.SettingType,
	_ bool, _ ...bunex.SelectQueryOption) (*entity.Setting, error) {
	if typ == base.SettingTypeSchedJob {
		if job, ok := r.jobs[r.last]; ok {
			return job, nil
		}
		return nil, hperrors.NewNotFound("Sched job")
	}
	setting, ok := r.byType[typ]
	if !ok {
		return nil, hperrors.NewNotFound("Setting")
	}
	r.last = setting.ID
	return setting, nil
}

func (r *settingsOfJobs) Update(_ context.Context, _ database.IDB, setting *entity.Setting,
	_ ...bunex.UpdateQueryOption) error {
	r.updated = append(r.updated, setting.ID)
	return nil
}

func dailySetting(t *testing.T, id, name string, data entity.SettingData) *entity.Setting {
	t.Helper()
	setting := &entity.Setting{ID: id, Name: name, Type: data.GetType()}
	if err := setting.SetData(data); err != nil {
		t.Fatal(err)
	}
	return setting
}

func jobOf(t *testing.T, id string, schedule entity.SchedJobSchedule) *entity.Setting {
	t.Helper()
	schedule.LastSchedTime = schedule.InitialTime.AddDate(0, 0, 6)
	schedule.LastInitialTime = schedule.InitialTime
	setting := &entity.Setting{ID: id, Type: base.SettingTypeSchedJob}
	if err := setting.SetData(&entity.SchedJob{JobType: base.SchedJobTypeSystemCleanup, Schedule: &schedule}); err != nil {
		t.Fatal(err)
	}
	return setting
}

// The jobs still at their default move to their time of day in the new
// timezone - their setting's schedule and their job's, the runs made under
// the old one forgotten. One an administrator changed is left, and named.
func TestMoveDailyJobs(t *testing.T) {
	utcMidnight := time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)
	cleanupSchedule := entity.SchedJobSchedule{Interval: daily, InitialTime: utcMidnight}
	backupSchedule := entity.SchedJobSchedule{Interval: daily, InitialTime: utcMidnight.Add(3 * time.Hour)}
	repo := &settingsOfJobs{
		byType: map[base.SettingType]*entity.Setting{
			base.SettingTypeSystemCleanup: dailySetting(t, "cleanup", "System cleanup",
				&entity.SystemCleanup{Schedule: cleanupSchedule}),
			base.SettingTypeSystemBackup: dailySetting(t, "backup", "System backup",
				&entity.SystemBackup{Schedule: backupSchedule}),
			base.SettingTypeSSLRenewal: dailySetting(t, "ssl", "SSL renewal",
				&entity.SSLRenewal{Schedule: entity.SchedJobSchedule{CronExpr: "0 4 * * *"}}),
		},
		jobs: map[string]*entity.Setting{
			"cleanup": jobOf(t, "cleanup-job", cleanupSchedule),
			"backup":  jobOf(t, "backup-job", backupSchedule),
		},
	}
	s := &service{settingRepo: repo}
	ny := newYork(t)

	moved, left, err := s.MoveDailyJobs(context.Background(), database.Tx{}, time.UTC, ny)

	assert.NoError(t, err)
	if assert.Len(t, moved, 1) {
		assert.Equal(t, "cleanup-job", moved[0].ID)
		job := moved[0].MustAsSchedJob().Schedule
		assert.Equal(t, time.Date(2026, 9, 30, 4, 0, 0, 0, time.UTC), job.InitialTime, "midnight there, that day")
		assert.Equal(t, daily, job.Interval)
		assert.True(t, job.LastSchedTime.IsZero(), "the runs made under the old schedule are forgotten")
		assert.Equal(t, 1, moved[0].UpdateVer)
	}
	cleanup := repo.byType[base.SettingTypeSystemCleanup].MustAsSystemCleanup()
	assert.Equal(t, time.Date(2026, 9, 30, 4, 0, 0, 0, time.UTC), cleanup.Schedule.InitialTime)
	assert.Equal(t, []string{"System backup", "SSL renewal"}, left)
	assert.Equal(t, []string{"cleanup", "cleanup-job"}, repo.updated)
	assert.Equal(t, utcMidnight.Add(3*time.Hour),
		repo.byType[base.SettingTypeSystemBackup].MustAsSystemBackup().Schedule.InitialTime, "left as it was")

	// Moved again, back: the job is at its default in the timezone it was moved to.
	moved, _, err = s.MoveDailyJobs(context.Background(), database.Tx{}, ny, time.UTC)
	assert.NoError(t, err)
	assert.Len(t, moved, 1)
	assert.Equal(t, utcMidnight, repo.byType[base.SettingTypeSystemCleanup].MustAsSystemCleanup().Schedule.InitialTime)
}
