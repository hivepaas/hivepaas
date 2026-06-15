package settingserviceimpl

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
	sysCleanupSettingName   = "System cleanup settings"
	sysCleanupJobName       = "System cleanup job"
	sysCleanupDefaultStatus = base.SettingStatusActive
	sysCleanupInterval      = timeutil.Duration(timeutil.Day) // daily
	sysCleanupMaxRetry      = 1
	sysCleanupRetryDelay    = timeutil.Duration(time.Second * 30)

	sysCleanupBackupRetention = timeutil.Duration(timeutil.Day * 30)

	dbObjectRetentionOfTasks          = timeutil.Duration(timeutil.Day * 90)
	dbObjectRetentionOfSysErrors      = timeutil.Duration(timeutil.Day * 90)
	dbObjectRetentionOfDeployments    = timeutil.Duration(timeutil.Day * 90)
	dbObjectRetentionOfDeletedObjects = timeutil.Duration(timeutil.Day * 90)

	sysCleanupRepoCacheRetention = timeutil.Duration(timeutil.Day * 10)
)

func (s *service) initDefaultSystemCleanup(
	ctx context.Context,
	db database.IDB,
	timeNow time.Time,
) (err error) {
	// Cleanup settings
	cleanupSetting := &entity.Setting{
		ID:        gofn.Must(ulid.NewStringULID()),
		Scope:     base.ObjectScopeGlobal,
		Type:      base.SettingTypeSystemCleanup,
		Status:    sysCleanupDefaultStatus,
		Name:      sysCleanupSettingName,
		Version:   entity.CurrentSystemCleanupVersion,
		CreatedAt: timeNow,
		UpdatedAt: timeNow,
	}
	cleanup := &entity.SystemCleanup{
		Schedule: entity.SchedJobSchedule{
			Interval:    sysCleanupInterval,
			InitialTime: time.Date(timeNow.Year(), timeNow.Month(), timeNow.Day(), 0, 0, 0, 0, time.UTC),
		},
		DBObjectRetention: entity.DBObjectRetention{
			Enabled:        true,
			Tasks:          dbObjectRetentionOfTasks,
			SysErrors:      dbObjectRetentionOfSysErrors,
			Deployments:    dbObjectRetentionOfDeployments,
			DeletedObjects: dbObjectRetentionOfDeletedObjects,
		},
		ClusterCleanup: entity.SystemClusterCleanup{
			Enabled:         true,
			PruneImages:     true,
			PruneVolumes:    true,
			PruneNetworks:   true,
			PruneContainers: true,
		},
		BackupCleanup: entity.SystemBackupCleanup{
			Enabled:              true,
			LocalBackupRetention: sysCleanupBackupRetention,
			CloudBackupRetention: sysCleanupBackupRetention,
		},
		CacheCleanup: entity.SystemCacheCleanup{
			Enabled:            true,
			RepoCacheRetention: sysCleanupRepoCacheRetention,
		},
		FileCleanup: entity.SystemFileCleanup{
			Enabled: true,
		},
		Notification: &entity.BaseEventNotification{
			SuccessUseDefault: true,
			FailureUseDefault: true,
		},
	}
	cleanupSetting.MustSetData(cleanup)

	// Cleanup job
	jobSetting := &entity.Setting{
		ID:        gofn.Must(ulid.NewStringULID()),
		Scope:     base.ObjectScopeGlobal,
		Type:      base.SettingTypeSchedJob,
		Kind:      string(base.SchedJobTypeSystemCleanup),
		Status:    sysCleanupDefaultStatus,
		Name:      sysCleanupJobName,
		Version:   entity.CurrentSchedJobVersion,
		CreatedAt: timeNow,
		UpdatedAt: timeNow,
	}
	schedJob := &entity.SchedJob{
		JobType:       base.SchedJobTypeSystemCleanup,
		Schedule:      &cleanup.Schedule,
		TargetSetting: entity.ObjectID{ID: cleanupSetting.ID},
		MaxRetry:      sysCleanupMaxRetry,
		RetryDelay:    sysCleanupRetryDelay,
		Notification:  cleanup.Notification,
	}
	jobSetting.MustSetData(schedJob)

	// Save the objects in DB
	err = s.settingRepo.InsertMulti(ctx, db, []*entity.Setting{cleanupSetting, jobSetting})
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}
