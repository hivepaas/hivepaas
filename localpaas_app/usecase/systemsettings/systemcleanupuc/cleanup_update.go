package systemcleanupuc

import (
	"context"
	"errors"
	"time"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/ulid"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/systemsettings/systemcleanupuc/systemcleanupdto"
)

const (
	currentSettingType   = base.SettingTypeSystemCleanup
	cleanupSettingName   = "System cleanup settings"
	cleanupJobName       = "System cleanup job"
	cleanupJobMaxRetry   = 1
	cleanupJobRetryDelay = timeutil.Duration(time.Second * 30)
)

func (uc *UC) UpdateSystemCleanup(
	ctx context.Context,
	auth *basedto.Auth,
	req *systemcleanupdto.UpdateSystemCleanupReq,
) (*systemcleanupdto.UpdateSystemCleanupResp, error) {
	req.Type = currentSettingType
	updateData := &updateSettingData{
		NewCleanup: req.ToEntity(),
	}
	persistingData := &persistingSettingData{}

	_, err := uc.UpdateSetting(ctx, &req.UpdateSettingReq, &settings.UpdateSettingData{
		VerifyingName: cleanupSettingName,
		Load: func(
			ctx context.Context,
			db database.Tx,
			data *settings.UpdateSettingData,
		) error {
			updateData.UpdateSettingData = data
			return uc.loadSettingData(ctx, db, req, updateData)
		},
		PrepareUpdate: func(
			ctx context.Context,
			db database.Tx,
			data *settings.UpdateSettingData,
			pData *settings.PersistingSettingData,
		) error {
			persistingData.PersistingSettingData = pData
			return uc.preparePersistingData(req, updateData, persistingData)
		},
		AfterPersisting: func(
			ctx context.Context,
			db database.Tx,
			data *settings.UpdateSettingData,
			pData *settings.PersistingSettingData,
		) error {
			return uc.postPersisting(ctx, db, updateData, persistingData)
		},
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &systemcleanupdto.UpdateSystemCleanupResp{}, nil
}

type updateSettingData struct {
	*settings.UpdateSettingData
	NewCleanup         *entity.SystemCleanup
	JobSetting         *entity.Setting
	JobScheduleChanges bool
}

type persistingSettingData struct {
	*settings.PersistingSettingData
	JobSetting *entity.Setting
}

func (uc *UC) loadSettingData(
	ctx context.Context,
	db database.Tx,
	req *systemcleanupdto.UpdateSystemCleanupReq,
	data *updateSettingData,
) error {
	cleanupSetting, err := uc.SettingRepo.GetSingle(ctx, db, req.Scope, base.SettingTypeSystemCleanup, false,
		bunex.SelectFor("UPDATE OF setting"),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}
	data.Setting = cleanupSetting

	cleanup, err := cleanupSetting.AsSystemCleanup()
	if err != nil {
		return apperrors.Wrap(err)
	}
	data.JobScheduleChanges = cleanup.ScheduleInterval != data.NewCleanup.ScheduleInterval ||
		cleanup.ScheduleFrom != data.NewCleanup.ScheduleFrom

	// Load cron job of the cleanup
	jobSetting, err := uc.SettingRepo.GetSingle(ctx, db, req.Scope, base.SettingTypeCronJob, false,
		bunex.SelectWhere("setting.data->'targetSetting'->>'id' = ?", cleanupSetting.ID),
		bunex.SelectFor("UPDATE OF setting"),
	)
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return apperrors.Wrap(err)
	}
	if jobSetting == nil {
		timeNow := timeutil.NowUTC()
		jobSetting = &entity.Setting{
			ID:        gofn.Must(ulid.NewStringULID()),
			Scope:     req.Scope.ScopeType(),
			Type:      base.SettingTypeCronJob,
			Status:    base.SettingStatusActive,
			Name:      cleanupJobName,
			Version:   entity.CurrentCronJobVersion,
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
		}
		cronJob := &entity.CronJob{
			CronType:      base.CronJobTypeSystemCleanup,
			Schedule:      &entity.CronJobSchedule{},
			TargetSetting: entity.ObjectID{ID: cleanupSetting.ID},
			MaxRetry:      cleanupJobMaxRetry,
			RetryDelay:    cleanupJobRetryDelay,
		}
		jobSetting.MustSetData(cronJob)
	}
	data.JobSetting = jobSetting

	return nil
}

func (uc *UC) preparePersistingData(
	req *systemcleanupdto.UpdateSystemCleanupReq,
	updateData *updateSettingData,
	persistingData *persistingSettingData,
) error {
	// Set new cleanup settings
	persistingData.Setting.Status = req.Status
	err := persistingData.Setting.SetData(updateData.NewCleanup)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Update cleanup job
	jobSetting := updateData.JobSetting
	jobSetting.Status = gofn.If(persistingData.Setting.Status == base.SettingStatusActive,
		base.SettingStatusActive, base.SettingStatusDisabled)
	jobSetting.Kind = string(base.CronJobTypeSystemCleanup)
	persistingData.JobSetting = jobSetting

	cleanupJob := jobSetting.MustAsCronJob()
	cleanupJob.Schedule.Interval = updateData.NewCleanup.ScheduleInterval
	cleanupJob.Schedule.InitialTime = updateData.NewCleanup.ScheduleFrom
	cleanupJob.Schedule.OnChange(updateData.JobScheduleChanges) // call this to handle if the schedule changes
	cleanupJob.Notification = updateData.NewCleanup.Notification
	jobSetting.MustSetData(cleanupJob)

	return nil
}

func (uc *UC) postPersisting(
	ctx context.Context,
	db database.Tx,
	updateData *updateSettingData,
	persistingData *persistingSettingData,
) error {
	// Persist the cron job updates
	err := uc.SettingRepo.Update(ctx, db, persistingData.JobSetting)
	if err != nil {
		return apperrors.Wrap(err)
	}

	err = uc.taskQueue.ScheduleTasksForCronJob(ctx, db, updateData.JobSetting, updateData.JobScheduleChanges)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}
