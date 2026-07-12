package settingserviceimpl

import (
	"context"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
)

const (
	sysBackupSettingName       = "System backup settings"
	sysBackupJobName           = "System backup job"
	sysBackupDefaultStatus     = base.SettingStatusDisabled      // Default to Disabled
	sysBackupInterval          = timeutil.Duration(timeutil.Day) // daily
	sysBackupMaxRetry          = 1
	sysBackupRetryDelay        = timeutil.Duration(time.Second * 60)
	sysBackupCompressionFormat = base.FileCompressionFormatGzip
	sysBackupDeletedObjects    = true
)

func (s *service) initDefaultSystemBackup(
	ctx context.Context,
	db database.IDB,
	timeNow time.Time,
) (err error) {
	// Backup settings
	backupSetting := &entity.Setting{
		ID:        gofn.Must(ulid.NewStringULID()),
		Scope:     base.ObjectScopeGlobal,
		Type:      base.SettingTypeSystemBackup,
		Status:    sysBackupDefaultStatus,
		Name:      sysBackupSettingName,
		Version:   entity.CurrentSystemBackupVersion,
		CreatedAt: timeNow,
		UpdatedAt: timeNow,
	}
	backup := &entity.SystemBackup{
		Schedule: entity.SchedJobSchedule{
			Interval:    sysBackupInterval,
			InitialTime: time.Date(timeNow.Year(), timeNow.Month(), timeNow.Day(), 0, 30, 0, 0, time.UTC),
		},
		DBBackupConfig: entity.SystemBackupDBConfig{
			BackupDeletedObjects: sysBackupDeletedObjects,
		},
		Compression: entity.SystemBackupCompression{
			Format: sysBackupCompressionFormat,
		},
		Encryption: entity.SystemBackupEncryption{
			Format: base.FileEncryptionNone,
		},
		Notification: &entity.BaseEventNotification{
			SuccessUseDefault: true,
			FailureUseDefault: true,
		},
	}
	backupSetting.MustSetData(backup)

	// Backup job
	jobSetting := &entity.Setting{
		ID:        gofn.Must(ulid.NewStringULID()),
		Scope:     base.ObjectScopeGlobal,
		Type:      base.SettingTypeSchedJob,
		Kind:      string(base.SchedJobTypeSystemBackup),
		Status:    sysBackupDefaultStatus,
		Name:      sysBackupJobName,
		Version:   entity.CurrentSchedJobVersion,
		CreatedAt: timeNow,
		UpdatedAt: timeNow,
	}
	schedJob := &entity.SchedJob{
		JobType:       base.SchedJobTypeSystemBackup,
		Schedule:      &backup.Schedule,
		TargetSetting: entity.ObjectID{ID: backupSetting.ID},
		MaxRetry:      sysBackupMaxRetry,
		RetryDelay:    sysBackupRetryDelay,
		Notification:  backup.Notification,
	}
	jobSetting.MustSetData(schedJob)

	// Save the objects in DB
	err = s.settingRepo.InsertMulti(ctx, db, []*entity.Setting{backupSetting, jobSetting})
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}
