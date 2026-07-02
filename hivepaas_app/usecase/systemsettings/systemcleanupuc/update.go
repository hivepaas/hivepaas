package systemcleanupuc

import (
	"context"
	"errors"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/systemsettings/systemcleanupuc/systemcleanupdto"
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
		return nil, apperrors.New(err)
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
		return apperrors.New(err)
	}
	data.Setting = cleanupSetting

	cleanup, err := cleanupSetting.AsSystemCleanup()
	if err != nil {
		return apperrors.New(err)
	}
	data.JobScheduleChanges = !cleanup.Schedule.Equal(&data.NewCleanup.Schedule)

	// Load sched job of the cleanup
	jobSetting, err := uc.SettingRepo.GetSingle(ctx, db, req.Scope, base.SettingTypeSchedJob, false,
		bunex.SelectWhere("setting.data->'targetSetting'->>'id' = ?", cleanupSetting.ID),
		bunex.SelectFor("UPDATE OF setting"),
	)
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return apperrors.New(err)
	}
	if jobSetting == nil {
		timeNow := timeutil.NowUTC()
		jobSetting = &entity.Setting{
			ID:        gofn.Must(ulid.NewStringULID()),
			Scope:     req.Scope.ScopeType(),
			Type:      base.SettingTypeSchedJob,
			Status:    base.SettingStatusActive,
			Name:      cleanupJobName,
			Version:   entity.CurrentSchedJobVersion,
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
		}
		schedJob := &entity.SchedJob{
			JobType:       base.SchedJobTypeSystemCleanup,
			Schedule:      &entity.SchedJobSchedule{},
			TargetSetting: entity.ObjectID{ID: cleanupSetting.ID},
			MaxRetry:      cleanupJobMaxRetry,
			RetryDelay:    cleanupJobRetryDelay,
		}
		jobSetting.MustSetData(schedJob)
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
		return apperrors.New(err)
	}

	// Update cleanup job
	jobSetting := updateData.JobSetting
	jobSetting.Status = gofn.If(persistingData.Setting.Status == base.SettingStatusActive,
		base.SettingStatusActive, base.SettingStatusDisabled)
	jobSetting.Kind = string(base.SchedJobTypeSystemCleanup)
	persistingData.JobSetting = jobSetting

	cleanupJob := jobSetting.MustAsSchedJob()
	cleanupJob.Schedule = &updateData.NewCleanup.Schedule
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
	// Persist the sched job updates
	err := uc.SettingRepo.Update(ctx, db, persistingData.JobSetting)
	if err != nil {
		return apperrors.New(err)
	}

	err = uc.taskQueue.ScheduleTasksForSchedJob(ctx, db, updateData.JobSetting, updateData.JobScheduleChanges)
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}
