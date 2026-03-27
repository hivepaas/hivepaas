package settingservice

import (
	"context"
	"time"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/ulid"
)

const (
	sslRenewalSettingName   = "SSL renewal settings"
	sslRenewalJobName       = "SSL renewal job"
	sslRenewalDefaultStatus = base.SettingStatusActive
	sslRenewalInterval      = timeutil.Duration(time.Hour * 24) // daily
	sslRenewalMaxRetry      = 1
	sslRenewalRetryDelay    = timeutil.Duration(time.Second * 60)
)

func (s *settingService) initDefaultSSLRenewal(
	ctx context.Context,
	db database.IDB,
	timeNow time.Time,
) (err error) {
	// Renewal settings
	renewalSetting := &entity.Setting{
		ID:        gofn.Must(ulid.NewStringULID()),
		Scope:     base.SettingScopeGlobal,
		Type:      base.SettingTypeSSLRenewal,
		Status:    sslRenewalDefaultStatus,
		Name:      sslRenewalSettingName,
		Version:   entity.CurrentSSLRenewalVersion,
		CreatedAt: timeNow,
		UpdatedAt: timeNow,
	}
	renewal := &entity.SSLRenewal{
		ScheduleInterval: sslRenewalInterval,
		ScheduleFrom:     timeNow.Truncate(sslRenewalInterval.ToDuration()),
	}
	renewalSetting.MustSetData(renewal)

	// Renewal job
	jobSetting := &entity.Setting{
		ID:        gofn.Must(ulid.NewStringULID()),
		Scope:     base.SettingScopeGlobal,
		Type:      base.SettingTypeCronJob,
		Kind:      string(base.CronJobTypeSSLRenewal),
		Status:    sslRenewalDefaultStatus,
		Name:      sslRenewalJobName,
		Version:   entity.CurrentCronJobVersion,
		CreatedAt: timeNow,
		UpdatedAt: timeNow,
	}
	cronJob := &entity.CronJob{
		CronType: base.CronJobTypeSSLRenewal,
		Schedule: &entity.CronJobSchedule{
			Interval:    renewal.ScheduleInterval,
			InitialTime: renewal.ScheduleFrom,
		},
		TargetSetting: entity.ObjectID{ID: renewalSetting.ID},
		MaxRetry:      sslRenewalMaxRetry,
		RetryDelay:    sslRenewalRetryDelay,
	}
	jobSetting.MustSetData(cronJob)

	// Save the objects in DB
	err = s.settingRepo.InsertMulti(ctx, db, []*entity.Setting{renewalSetting, jobSetting})
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}
