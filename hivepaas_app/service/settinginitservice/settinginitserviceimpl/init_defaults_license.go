package settinginitserviceimpl

import (
	"context"
	"errors"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
)

const (
	licenseRefreshJobName    = "License refresh job"
	licenseRefreshInterval   = timeutil.Duration(timeutil.Day) // daily
	licenseRefreshMaxRetry   = 1
	licenseRefreshRetryDelay = timeutil.Duration(time.Hour)
)

// initDefaultLicense gives the license setting the daily refresh's schedule, and
// the refresh its job. It runs on every start rather than only when the setting
// is missing: an installation from before online refresh may have a key stored
// already, and no job. It changes nothing once both are there.
//
// A setting made here holds no key, which reads as no license, as no setting
// does; the first key installed is then recorded as an update of it.
//
// The job has no notification: a check that failed says so on System > License,
// and the next day's tries again.
func (s *service) initDefaultLicense(ctx context.Context, db database.Tx, timeNow time.Time) error {
	global := entity.NewObjectScopeGlobal()
	setting, err := s.settingRepo.GetSingle(ctx, db, global, base.SettingTypeLicense, false,
		bunex.SelectFor("UPDATE OF setting"))
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return hperrors.Wrap(err)
	}
	created := setting == nil
	if created {
		setting = &entity.Setting{
			ID:        gofn.Must(ulid.NewStringULID()),
			Scope:     base.ObjectScopeGlobal,
			Type:      base.SettingTypeLicense,
			Status:    base.SettingStatusActive,
			Name:      entity.LicenseSettingName,
			Version:   entity.CurrentLicenseSettingsVersion,
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
		}
		setting.MustSetData(&entity.LicenseSettings{})
	}
	settings, err := setting.AsLicenseSettings()
	if err != nil {
		return hperrors.Wrap(err)
	}

	scheduleSet := settings.Refresh.Schedule.InitialTime.IsZero()
	if scheduleSet {
		settings.Refresh.Schedule = entity.SchedJobSchedule{
			Interval:    licenseRefreshInterval,
			InitialTime: dailyJobOf(base.SettingTypeLicense).startIn(timeNow, timeutil.Location()),
		}
	}
	switch {
	case created:
		setting.MustSetData(settings)
		if err = s.settingRepo.Insert(ctx, db, setting); err != nil {
			return hperrors.Wrap(err)
		}
	case scheduleSet:
		if err = s.saveSetting(ctx, db, setting, settings); err != nil {
			return hperrors.Wrap(err)
		}
	}

	_, err = s.settingRepo.GetSingle(ctx, db, global, base.SettingTypeSchedJob, false,
		bunex.SelectWhere("setting.kind = ?", base.SchedJobTypeLicenseRefresh))
	if err == nil {
		return nil
	}
	if !errors.Is(err, hperrors.ErrNotFound) {
		return hperrors.Wrap(err)
	}
	return hperrors.Wrap(s.settingRepo.Insert(ctx, db, licenseRefreshJob(setting.ID, &settings.Refresh.Schedule,
		timeNow)))
}

// licenseRefreshJob is the daily refresh's job, on the license setting's
// schedule. The job's runs check whether the refresh is switched off, so the
// job is active whatever the config says.
func licenseRefreshJob(licenseSettingID string, schedule *entity.SchedJobSchedule,
	timeNow time.Time) *entity.Setting {
	jobSetting := &entity.Setting{
		ID:        gofn.Must(ulid.NewStringULID()),
		Scope:     base.ObjectScopeGlobal,
		Type:      base.SettingTypeSchedJob,
		Kind:      string(base.SchedJobTypeLicenseRefresh),
		Status:    base.SettingStatusActive,
		Name:      licenseRefreshJobName,
		Version:   entity.CurrentSchedJobVersion,
		CreatedAt: timeNow,
		UpdatedAt: timeNow,
	}
	jobSchedule := *schedule
	jobSetting.MustSetData(&entity.SchedJob{
		JobType:       base.SchedJobTypeLicenseRefresh,
		Schedule:      &jobSchedule,
		TargetSetting: entity.ObjectID{ID: licenseSettingID},
		MaxRetry:      licenseRefreshMaxRetry,
		RetryDelay:    licenseRefreshRetryDelay,
	})
	return jobSetting
}
