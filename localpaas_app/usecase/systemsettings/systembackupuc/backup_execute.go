package systembackupuc

import (
	"context"
	"time"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/usecase/systemsettings/systembackupuc/systembackupdto"
)

func (uc *UC) ExecuteSystemBackup(
	ctx context.Context,
	auth *basedto.Auth,
	req *systembackupdto.ExecuteSystemBackupReq,
) (*systembackupdto.ExecuteSystemBackupResp, error) {
	req.Type = currentSettingType
	_, jobSetting, err := uc.getBackupSettingAndJob(ctx, uc.DB, req.Scope, true, false)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	task, err := uc.cronJobService.CreateCronJobTask(jobSetting, time.Time{}, timeutil.NowUTC())
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	err = uc.taskRepo.Insert(ctx, uc.DB, task)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	err = uc.taskQueue.ScheduleTask(ctx, task)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &systembackupdto.ExecuteSystemBackupResp{
		Data: &systembackupdto.ExecuteSystemBackupDataResp{
			Task: &basedto.ObjectIDResp{ID: task.ID},
		},
	}, nil
}

func (uc *UC) getBackupSettingAndJob(
	ctx context.Context,
	db database.IDB,
	scope *base.SettingScope,
	requireSettingActive bool,
	requireJobActive bool,
) (backup *entity.Setting, job *entity.Setting, err error) {
	backup, err = uc.SettingRepo.GetSingle(ctx, db, scope, currentSettingType, requireSettingActive)
	if err != nil {
		return nil, nil, apperrors.Wrap(err)
	}

	// Load cron job of the backup
	job, err = uc.SettingRepo.GetSingle(ctx, db, scope, base.SettingTypeCronJob, requireJobActive,
		bunex.SelectWhere("setting.data->'targetSetting'->>'id' = ?", backup.ID),
	)
	if err != nil {
		return nil, nil, apperrors.Wrap(err)
	}

	return backup, job, nil
}
