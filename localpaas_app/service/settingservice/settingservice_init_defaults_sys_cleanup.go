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
	sysCleanupSettingName = "System cleanup settings"
	sysCleanupJobName     = "System cleanup job"
	sysCleanupInterval    = timeutil.Duration(time.Hour * 24) // daily
	sysCleanupMaxRetry    = 1
	sysCleanupRetryDelay  = timeutil.Duration(time.Second * 30)

	dbObjectRetentionOfTasks          = timeutil.Duration(time.Hour * 24 * 180) // 180 days
	dbObjectRetentionOfSysErrors      = timeutil.Duration(time.Hour * 24 * 180)
	dbObjectRetentionOfDeployments    = timeutil.Duration(time.Hour * 24 * 180)
	dbObjectRetentionOfDeletedObjects = timeutil.Duration(time.Hour * 24 * 180)
)

func (s *settingService) initDefaultSystemCleanup(
	ctx context.Context,
	db database.IDB,
	timeNow time.Time,
) (err error) {
	// Cleanup settings
	cleanupSetting := &entity.Setting{
		ID:        gofn.Must(ulid.NewStringULID()),
		Type:      base.SettingTypeSystemCleanup,
		Status:    base.SettingStatusActive,
		Name:      sysCleanupSettingName,
		Version:   entity.CurrentSystemCleanupVersion,
		CreatedAt: timeNow,
		UpdatedAt: timeNow,
	}
	cleanup := &entity.SystemCleanup{
		ScheduleInterval: sysCleanupInterval,
		ScheduleFrom:     timeNow.Truncate(sysCleanupInterval.ToDuration()),
		DBObjectRetention: &entity.DBObjectRetention{
			Tasks:          dbObjectRetentionOfTasks,
			SysErrors:      dbObjectRetentionOfSysErrors,
			Deployments:    dbObjectRetentionOfDeployments,
			DeletedObjects: dbObjectRetentionOfDeletedObjects,
		},
		ClusterCleanup: &entity.ClusterCleanup{
			PruneImages:     true,
			PruneVolumes:    true,
			PruneNetworks:   true,
			PruneContainers: true,
		},
	}
	cleanupSetting.MustSetData(cleanup)

	// Cleanup job
	jobSetting := &entity.Setting{
		ID:        gofn.Must(ulid.NewStringULID()),
		Type:      base.SettingTypeCronJob,
		Kind:      string(base.CronJobTypeSystemCleanup),
		Status:    base.SettingStatusActive,
		Name:      sysCleanupJobName,
		Version:   entity.CurrentCronJobVersion,
		CreatedAt: timeNow,
		UpdatedAt: timeNow,
	}
	cronJob := &entity.CronJob{
		CronType: base.CronJobTypeSystemCleanup,
		Schedule: &entity.CronJobSchedule{
			Interval:    cleanup.ScheduleInterval,
			InitialTime: cleanup.ScheduleFrom,
		},
		TargetSetting: entity.ObjectID{ID: cleanupSetting.ID},
		MaxRetry:      sysCleanupMaxRetry,
		RetryDelay:    sysCleanupRetryDelay,
		Notification:  cleanup.Notification,
	}
	jobSetting.MustSetData(cronJob)

	// Save the objects in DB
	err = s.settingRepo.InsertMulti(ctx, db, []*entity.Setting{cleanupSetting, jobSetting})
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}
