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

// licenseSettings holds the license setting and the refresh's job, if any.
type licenseSettings struct {
	repository.SettingRepo
	license  *entity.Setting
	job      *entity.Setting
	inserted []*entity.Setting
	updated  []string
}

func (r *licenseSettings) GetSingle(_ context.Context, _ database.IDB, _ *entity.ObjectScope, typ base.SettingType,
	_ bool, _ ...bunex.SelectQueryOption) (*entity.Setting, error) {
	found := r.license
	if typ == base.SettingTypeSchedJob {
		found = r.job
	}
	if found == nil {
		return nil, hperrors.NewNotFound("Setting")
	}
	return found, nil
}

func (r *licenseSettings) Insert(_ context.Context, _ database.IDB, setting *entity.Setting,
	_ ...bunex.InsertQueryOption) error {
	r.inserted = append(r.inserted, setting)
	if setting.Type == base.SettingTypeSchedJob {
		r.job = setting
	} else {
		r.license = setting
	}
	return nil
}

func (r *licenseSettings) Update(_ context.Context, _ database.IDB, setting *entity.Setting,
	_ ...bunex.UpdateQueryOption) error {
	r.updated = append(r.updated, setting.ID)
	return nil
}

var licenseInitNow = time.Date(2026, 11, 2, 15, 0, 0, 0, time.UTC)

// A new installation gets a license setting with no key, which reads as no
// license, and the refresh's job: daily, at 03:30 there, with one retry an hour
// later and no notification.
func TestInitDefaultLicenseMakesTheSettingAndTheJob(t *testing.T) {
	repo := &licenseSettings{}
	s := &service{settingRepo: repo}

	assert.NoError(t, s.initDefaultLicense(context.Background(), database.Tx{}, licenseInitNow))

	if !assert.Len(t, repo.inserted, 2) {
		return
	}
	license := repo.license
	assert.Equal(t, base.SettingTypeLicense, license.Type)
	assert.Equal(t, base.ObjectScopeGlobal, license.Scope)
	assert.Equal(t, base.SettingStatusActive, license.Status)
	assert.Equal(t, entity.LicenseSettingName, license.Name)
	assert.Equal(t, entity.CurrentLicenseSettingsVersion, license.Version)
	settings := license.MustAsLicenseSettings()
	assert.Empty(t, settings.Key)
	schedule := entity.SchedJobSchedule{Interval: timeutil.Duration(timeutil.Day),
		InitialTime: time.Date(2026, 11, 2, 3, 30, 0, 0, time.UTC)}
	assert.Equal(t, schedule, settings.Refresh.Schedule)

	job := repo.job
	assert.Equal(t, base.SettingTypeSchedJob, job.Type)
	assert.Equal(t, string(base.SchedJobTypeLicenseRefresh), job.Kind)
	assert.Equal(t, base.SettingStatusActive, job.Status)
	schedJob := job.MustAsSchedJob()
	assert.Equal(t, base.SchedJobTypeLicenseRefresh, schedJob.JobType)
	assert.Equal(t, license.ID, schedJob.TargetSetting.ID)
	assert.Equal(t, &schedule, schedJob.Schedule)
	assert.Equal(t, 1, schedJob.MaxRetry)
	assert.Equal(t, timeutil.Duration(time.Hour), schedJob.RetryDelay)
	assert.Nil(t, schedJob.Notification)
}

// A license stored before online refresh keeps its key and gets the schedule,
// and its job.
func TestInitDefaultLicenseAddsTheJobToAStoredLicense(t *testing.T) {
	license := &entity.Setting{ID: "LICENSE", Type: base.SettingTypeLicense, Name: entity.LicenseSettingName}
	license.MustSetData(&entity.LicenseSettings{Key: "the-key"})
	repo := &licenseSettings{license: license}
	s := &service{settingRepo: repo}

	assert.NoError(t, s.initDefaultLicense(context.Background(), database.Tx{}, licenseInitNow))

	assert.Equal(t, []string{"LICENSE"}, repo.updated)
	assert.Equal(t, 1, license.UpdateVer)
	settings := license.MustAsLicenseSettings()
	assert.Equal(t, "the-key", settings.Key)
	assert.Equal(t, time.Date(2026, 11, 2, 3, 30, 0, 0, time.UTC), settings.Refresh.Schedule.InitialTime)
	if assert.Len(t, repo.inserted, 1) {
		assert.Equal(t, "LICENSE", repo.job.MustAsSchedJob().TargetSetting.ID)
	}
}

// Once both are there, nothing is written: it runs on every start.
func TestInitDefaultLicenseIsIdempotent(t *testing.T) {
	repo := &licenseSettings{}
	s := &service{settingRepo: repo}
	assert.NoError(t, s.initDefaultLicense(context.Background(), database.Tx{}, licenseInitNow))
	schedule := repo.license.MustAsLicenseSettings().Refresh.Schedule

	assert.NoError(t, s.initDefaultLicense(context.Background(), database.Tx{}, licenseInitNow.Add(48*time.Hour)))

	assert.Len(t, repo.inserted, 2)
	assert.Empty(t, repo.updated)
	assert.Equal(t, schedule, repo.license.MustAsLicenseSettings().Refresh.Schedule)
}

// The license's daily job moves with the others: its schedule is the license
// setting's refresh's.
func TestLicenseRefreshIsADailyJob(t *testing.T) {
	job := dailyJobOf(base.SettingTypeLicense)
	ny := newYork(t)
	// 02:00 UTC on 6 October is 22:00 on the 5th in New York.
	assert.Equal(t, time.Date(2026, 10, 5, 7, 30, 0, 0, time.UTC),
		job.startIn(time.Date(2026, 10, 6, 2, 0, 0, 0, time.UTC), ny), "03:30 of the 5th there")

	schedule := entity.SchedJobSchedule{Interval: daily, InitialTime: time.Date(2026, 9, 30, 3, 30, 0, 0, time.UTC)}
	license := dailySetting(t, "license", entity.LicenseSettingName,
		&entity.LicenseSettings{Key: "the-key", Refresh: entity.LicenseRefresh{Schedule: schedule}})
	repo := &settingsOfJobs{
		byType: map[base.SettingType]*entity.Setting{base.SettingTypeLicense: license},
		jobs:   map[string]*entity.Setting{"license": jobOf(t, "license-job", schedule)},
	}

	moved, left, err := (&service{settingRepo: repo}).MoveDailyJobs(context.Background(), database.Tx{}, time.UTC, ny)

	assert.NoError(t, err)
	assert.Empty(t, left)
	if assert.Len(t, moved, 1) {
		assert.Equal(t, time.Date(2026, 9, 30, 7, 30, 0, 0, time.UTC), moved[0].MustAsSchedJob().Schedule.InitialTime)
	}
	settings := license.MustAsLicenseSettings()
	assert.Equal(t, "the-key", settings.Key, "the key is written back as it was")
	assert.Equal(t, time.Date(2026, 9, 30, 7, 30, 0, 0, time.UTC), settings.Refresh.Schedule.InitialTime)
}
