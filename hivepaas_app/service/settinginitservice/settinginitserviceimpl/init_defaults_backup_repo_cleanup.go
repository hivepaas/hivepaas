package settinginitserviceimpl

import (
	"context"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
)

const (
	backupRepoCleanupSettingName   = "Backup repo cleanup settings"
	backupRepoCleanupJobName       = "Backup repo cleanup job"
	backupRepoCleanupDefaultStatus = base.SettingStatusActive
	backupRepoCleanupInterval      = timeutil.Duration(timeutil.Day) // daily
	backupRepoCleanupMaxRetry      = 1
	backupRepoCleanupRetryDelay    = timeutil.Duration(time.Second * 60)
)

func (s *service) initDefaultBackupRepoCleanup(
	ctx context.Context,
	db database.IDB,
	timeNow time.Time,
) (err error) {
	// Cleanup settings
	cleanupSetting := &entity.Setting{
		ID:          gofn.Must(ulid.NewStringULID()),
		Type:        base.SettingTypeBackupRepoCleanup,
		Status:      backupRepoCleanupDefaultStatus,
		Name:        backupRepoCleanupSettingName,
		Inheritable: false,
		Version:     entity.CurrentBackupRepoCleanupVersion,
		CreatedAt:   timeNow,
		UpdatedAt:   timeNow,
	}
	cleanup := &entity.BackupRepoCleanup{
		Schedule: entity.SchedJobSchedule{
			Interval:    backupRepoCleanupInterval,
			InitialTime: time.Date(timeNow.Year(), timeNow.Month(), timeNow.Day(), 1, 30, 0, 0, time.UTC),
		},
		Notification: &entity.BaseEventNotification{},
	}
	cleanupSetting.MustSetData(cleanup)

	// Cleanup job
	jobSetting := &entity.Setting{
		ID:          gofn.Must(ulid.NewStringULID()),
		Type:        base.SettingTypeSchedJob,
		Kind:        string(base.SchedJobTypeBackupRepoCleanup),
		Status:      backupRepoCleanupDefaultStatus,
		Name:        backupRepoCleanupJobName,
		Inheritable: false,
		Version:     entity.CurrentSchedJobVersion,
		CreatedAt:   timeNow,
		UpdatedAt:   timeNow,
	}
	schedJob := &entity.SchedJob{
		JobType:       base.SchedJobTypeBackupRepoCleanup,
		Schedule:      &cleanup.Schedule,
		TargetSetting: entity.ObjectID{ID: cleanupSetting.ID},
		MaxRetry:      backupRepoCleanupMaxRetry,
		RetryDelay:    backupRepoCleanupRetryDelay,
		Notification:  cleanup.Notification,
	}
	jobSetting.MustSetData(schedJob)

	// Save the objects in DB
	err = s.settingRepo.InsertMulti(ctx, db, []*entity.Setting{cleanupSetting, jobSetting})
	if err != nil {
		return hperrors.Wrap(err)
	}

	return nil
}
