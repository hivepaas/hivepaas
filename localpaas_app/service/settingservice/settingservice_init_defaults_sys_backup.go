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
	sysBackupSettingName          = "System backup settings"
	sysBackupJobName              = "System backup job"
	sysBackupDefaultStatus        = base.SettingStatusDisabled        // Default to Disabled
	sysBackupInterval             = timeutil.Duration(time.Hour * 24) // daily
	sysBackupMaxRetry             = 1
	sysBackupRetryDelay           = timeutil.Duration(time.Second * 60)
	sysBackupCompression          = true
	sysBackupDeletedObjects       = true
	sysBackupLocalBackupRetention = timeutil.Duration(time.Hour * 24 * 30) // 30 days
)

func (s *settingService) initDefaultSystemBackup(
	ctx context.Context,
	db database.IDB,
	timeNow time.Time,
) (err error) {
	// Backup settings
	backupSetting := &entity.Setting{
		ID:        gofn.Must(ulid.NewStringULID()),
		Scope:     base.SettingScopeGlobal,
		Type:      base.SettingTypeSystemBackup,
		Status:    sysBackupDefaultStatus,
		Name:      sysBackupSettingName,
		Version:   entity.CurrentSystemBackupVersion,
		CreatedAt: timeNow,
		UpdatedAt: timeNow,
	}
	backup := &entity.SystemBackup{
		ScheduleInterval: sysBackupInterval,
		ScheduleFrom:     timeNow.Truncate(sysBackupInterval.ToDuration()),
		DBBackupConfig: &entity.DBBackupConfig{
			BackupDeletedObjects: sysBackupDeletedObjects,
		},
		Compression:          sysBackupCompression,
		LocalBackupRetention: sysBackupLocalBackupRetention,
		Notification: &entity.BaseEventNotification{
			SuccessUseDefault: true,
			FailureUseDefault: true,
		},
	}
	backupSetting.MustSetData(backup)

	// Backup job
	jobSetting := &entity.Setting{
		ID:        gofn.Must(ulid.NewStringULID()),
		Scope:     base.SettingScopeGlobal,
		Type:      base.SettingTypeCronJob,
		Kind:      string(base.CronJobTypeSystemBackup),
		Status:    sysBackupDefaultStatus,
		Name:      sysBackupJobName,
		Version:   entity.CurrentCronJobVersion,
		CreatedAt: timeNow,
		UpdatedAt: timeNow,
	}
	cronJob := &entity.CronJob{
		CronType: base.CronJobTypeSystemBackup,
		Schedule: &entity.CronJobSchedule{
			Interval:    backup.ScheduleInterval,
			InitialTime: backup.ScheduleFrom,
		},
		TargetSetting: entity.ObjectID{ID: backupSetting.ID},
		MaxRetry:      sysBackupMaxRetry,
		RetryDelay:    sysBackupRetryDelay,
		Notification:  backup.Notification,
	}
	jobSetting.MustSetData(cronJob)

	// Save the objects in DB
	err = s.settingRepo.InsertMulti(ctx, db, []*entity.Setting{backupSetting, jobSetting})
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}
