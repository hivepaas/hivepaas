package systembackupuc

import (
	"context"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/systemsettings/systembackupuc/systembackupdto"
)

func (uc *UC) ExecuteSystemBackup(
	ctx context.Context,
	auth *basedto.Auth,
	req *systembackupdto.ExecuteSystemBackupReq,
) (*systembackupdto.ExecuteSystemBackupResp, error) {
	req.Type = currentSettingType
	_, jobSetting, err := uc.getBackupSettingAndJob(ctx, uc.DB, req.Scope, true, false)
	if err != nil {
		return nil, apperrors.New(err)
	}

	task, err := uc.schedJobService.CreateSchedJobTask(jobSetting, time.Time{}, timeutil.NowUTC())
	if err != nil {
		return nil, apperrors.New(err)
	}

	err = uc.taskRepo.Insert(ctx, uc.DB, task)
	if err != nil {
		return nil, apperrors.New(err)
	}

	err = uc.taskQueue.ScheduleTask(ctx, task)
	if err != nil {
		return nil, apperrors.New(err)
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
	scope *base.ObjectScope,
	requireSettingActive bool,
	requireJobActive bool,
) (backup *entity.Setting, job *entity.Setting, err error) {
	backup, err = uc.SettingRepo.GetSingle(ctx, db, scope, currentSettingType, requireSettingActive)
	if err != nil {
		return nil, nil, apperrors.New(err)
	}

	// Load sched job of the backup
	job, err = uc.SettingRepo.GetSingle(ctx, db, scope, base.SettingTypeSchedJob, requireJobActive,
		bunex.SelectWhere("setting.data->'targetSetting'->>'id' = ?", backup.ID),
	)
	if err != nil {
		return nil, nil, apperrors.New(err)
	}

	return backup, job, nil
}
