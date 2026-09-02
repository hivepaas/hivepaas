package backuprepocleanupuc

import (
	"context"
	"errors"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/systemsettings/backuprepocleanupuc/backuprepocleanupdto"
)

const (
	currentSettingType   = base.SettingTypeBackupRepoCleanup
	cleanupSettingName   = "Backup repo cleanup settings"
	cleanupJobName       = "Backup repo cleanup job"
	cleanupJobMaxRetry   = 1
	cleanupJobRetryDelay = timeutil.Duration(time.Second * 60)
)

func (uc *UC) UpdateBackupRepoCleanup(
	ctx context.Context,
	auth *basedto.Auth,
	req *backuprepocleanupdto.UpdateBackupRepoCleanupReq,
) (*backuprepocleanupdto.UpdateBackupRepoCleanupResp, error) {
	req.Type = currentSettingType
	updateData := &updateSettingData{
		NewSettings: req.ToEntity(),
	}
	persistingData := &persistingSettingData{}

	_, err := uc.UpdateUniqueSetting(ctx, &req.UpdateUniqueSettingReq, &settings.UpdateUniqueSettingData{
		Name: cleanupSettingName,
		Load: func(
			ctx context.Context,
			db database.Tx,
			data *settings.UpdateUniqueSettingData,
		) error {
			updateData.UpdateUniqueSettingData = data
			return uc.loadSettingData(ctx, db, req, updateData)
		},
		PrepareUpdate: func(
			ctx context.Context,
			db database.Tx,
			data *settings.UpdateUniqueSettingData,
			pData *settings.PersistingSettingData,
		) error {
			persistingData.PersistingSettingData = pData
			return uc.preparePersistingData(req, updateData, persistingData)
		},
		AfterPersisting: func(
			ctx context.Context,
			db database.Tx,
			data *settings.UpdateUniqueSettingData,
			pData *settings.PersistingSettingData,
		) error {
			return uc.postPersisting(ctx, db, updateData, persistingData)
		},
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &backuprepocleanupdto.UpdateBackupRepoCleanupResp{}, nil
}

type updateSettingData struct {
	*settings.UpdateUniqueSettingData
	NewSettings        *entity.BackupRepoCleanup
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
	req *backuprepocleanupdto.UpdateBackupRepoCleanupReq,
	data *updateSettingData,
) error {
	cleanupSetting, err := uc.SettingRepo.GetSingle(ctx, db, req.Scope, base.SettingTypeBackupRepoCleanup, false,
		bunex.SelectFor("UPDATE OF setting"),
	)
	if err != nil {
		return hperrors.Wrap(err)
	}
	data.Setting = cleanupSetting

	cleanup, err := cleanupSetting.AsBackupRepoCleanup()
	if err != nil {
		return hperrors.Wrap(err)
	}
	data.JobScheduleChanges = !cleanup.Schedule.Equal(&data.NewSettings.Schedule)

	// Load sched job of the cleanup
	jobSetting, err := uc.SettingRepo.GetSingle(ctx, db, req.Scope, base.SettingTypeSchedJob, false,
		bunex.SelectWhere("setting.kind = ?", base.SchedJobTypeBackupRepoCleanup),
		bunex.SelectWhere("setting.data->'targetSetting'->>'id' = ?", cleanupSetting.ID),
		bunex.SelectFor("UPDATE OF setting"),
	)
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return hperrors.Wrap(err)
	}
	if jobSetting == nil {
		timeNow := timeutil.NowUTC()
		jobSetting = &entity.Setting{
			ID:          gofn.Must(ulid.NewStringULID()),
			Scope:       req.Scope.ScopeType,
			Type:        base.SettingTypeSchedJob,
			Kind:        string(base.SchedJobTypeBackupRepoCleanup),
			Status:      base.SettingStatusActive,
			Name:        cleanupJobName,
			Inheritable: true,
			Version:     entity.CurrentSchedJobVersion,
			CreatedAt:   timeNow,
			UpdatedAt:   timeNow,
		}
		schedJob := &entity.SchedJob{
			JobType:       base.SchedJobTypeBackupRepoCleanup,
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
	req *backuprepocleanupdto.UpdateBackupRepoCleanupReq,
	updateData *updateSettingData,
	persistingData *persistingSettingData,
) error {
	// Set new cleanup settings
	persistingData.Setting.Status = req.Status
	err := persistingData.Setting.SetData(updateData.NewSettings)
	if err != nil {
		return hperrors.Wrap(err)
	}

	// Update cleanup job
	jobSetting := updateData.JobSetting
	jobSetting.Status = gofn.If(persistingData.Setting.Status == base.SettingStatusActive,
		base.SettingStatusActive, base.SettingStatusDisabled)
	jobSetting.Kind = string(base.SchedJobTypeBackupRepoCleanup)
	persistingData.JobSetting = jobSetting

	cleanupJob := jobSetting.MustAsSchedJob()
	cleanupJob.Schedule = &updateData.NewSettings.Schedule
	cleanupJob.Notification = updateData.NewSettings.Notification
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
		return hperrors.Wrap(err)
	}

	err = uc.taskQueue.ScheduleTasksForSchedJob(ctx, db, updateData.JobSetting, updateData.JobScheduleChanges)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
