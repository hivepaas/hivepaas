package systembackupuc

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
	"github.com/localpaas/localpaas/localpaas_app/usecase/systemsettings/systembackupuc/systembackupdto"
)

const (
	currentSettingType  = base.SettingTypeSystemBackup
	backupSettingName   = "System backup settings"
	backupJobName       = "System backup job"
	backupJobMaxRetry   = 1
	backupJobRetryDelay = timeutil.Duration(time.Second * 60)
)

func (uc *UC) UpdateSystemBackup(
	ctx context.Context,
	auth *basedto.Auth,
	req *systembackupdto.UpdateSystemBackupReq,
) (*systembackupdto.UpdateSystemBackupResp, error) {
	req.Type = currentSettingType
	updateData := &updateSettingData{
		NewBackup: req.ToEntity(),
	}
	persistingData := &persistingSettingData{}

	_, err := uc.UpdateSetting(ctx, &req.UpdateSettingReq, &settings.UpdateSettingData{
		VerifyingName: backupSettingName,
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

	return &systembackupdto.UpdateSystemBackupResp{}, nil
}

type updateSettingData struct {
	*settings.UpdateSettingData
	NewBackup          *entity.SystemBackup
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
	req *systembackupdto.UpdateSystemBackupReq,
	data *updateSettingData,
) error {
	backupSetting, err := uc.SettingRepo.GetSingle(ctx, db, req.Scope, base.SettingTypeSystemBackup, false,
		bunex.SelectFor("UPDATE OF setting"),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}
	data.Setting = backupSetting

	backup, err := backupSetting.AsSystemBackup()
	if err != nil {
		return apperrors.Wrap(err)
	}
	data.JobScheduleChanges = backup.ScheduleInterval != data.NewBackup.ScheduleInterval ||
		backup.ScheduleFrom != data.NewBackup.ScheduleFrom

	// Load sched job of the backup
	jobSetting, err := uc.SettingRepo.GetSingle(ctx, db, req.Scope, base.SettingTypeSchedJob, false,
		bunex.SelectWhere("setting.data->'targetSetting'->>'id' = ?", backupSetting.ID),
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
			Type:      base.SettingTypeSchedJob,
			Status:    base.SettingStatusActive,
			Name:      backupJobName,
			Version:   entity.CurrentSchedJobVersion,
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
		}
		schedJob := &entity.SchedJob{
			JobType:       base.SchedJobTypeSystemBackup,
			Schedule:      &entity.SchedJobSchedule{},
			TargetSetting: entity.ObjectID{ID: backupSetting.ID},
			MaxRetry:      backupJobMaxRetry,
			RetryDelay:    backupJobRetryDelay,
		}
		jobSetting.MustSetData(schedJob)
	}
	data.JobSetting = jobSetting

	return nil
}

func (uc *UC) preparePersistingData(
	req *systembackupdto.UpdateSystemBackupReq,
	updateData *updateSettingData,
	persistingData *persistingSettingData,
) error {
	// Set new backup settings
	persistingData.Setting.Status = req.Status
	err := persistingData.Setting.SetData(updateData.NewBackup)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Update backup job
	jobSetting := updateData.JobSetting
	jobSetting.Status = gofn.If(persistingData.Setting.Status == base.SettingStatusActive,
		base.SettingStatusActive, base.SettingStatusDisabled)
	jobSetting.Kind = string(base.SchedJobTypeSystemBackup)
	persistingData.JobSetting = jobSetting

	backupJob := jobSetting.MustAsSchedJob()
	backupJob.Schedule.Interval = updateData.NewBackup.ScheduleInterval
	backupJob.Schedule.CronExpr = updateData.NewBackup.ScheduleCronExpr
	backupJob.Schedule.InitialTime = updateData.NewBackup.ScheduleFrom
	if err = backupJob.Schedule.IsValid(); err != nil {
		return apperrors.Wrap(err)
	}

	backupJob.Schedule.OnChange(updateData.JobScheduleChanges) // call this to handle if the schedule changes
	backupJob.Notification = updateData.NewBackup.Notification
	jobSetting.MustSetData(backupJob)

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
		return apperrors.Wrap(err)
	}

	err = uc.taskQueue.ScheduleTasksForSchedJob(ctx, db, updateData.JobSetting, updateData.JobScheduleChanges)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}
