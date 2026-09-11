package systembackupuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/systemsettings/systembackupuc/systembackupdto"
)

func (uc *UC) ExecuteSystemBackup(
	ctx context.Context,
	auth *basedto.Auth,
	req *systembackupdto.ExecuteSystemBackupReq,
) (*systembackupdto.ExecuteSystemBackupResp, error) {
	req.Type = currentSettingType
	backupSetting, jobSetting, err := uc.getBackupSettingAndJob(ctx, uc.DB, req.Scope, true, false)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	timeNow := timeutil.NowUTC()
	task, err := uc.schedJobService.CreateSchedJobTask(jobSetting, timeNow, timeNow)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	// The row and its record commit together, so there is no backup queued that
	// nothing accounts for and no record of one that never ran.
	err = transaction.Execute(ctx, uc.DB, func(db database.Tx) error {
		if err := uc.taskRepo.Insert(ctx, db, task); err != nil {
			return hperrors.Wrap(err)
		}
		return uc.recordBackupRun(ctx, db, auth, req.Scope, backupSetting, task)
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	// After the commit: handing the queue a task whose row rolled back would have
	// it run a backup nobody asked for.
	err = uc.taskQueue.ScheduleTask(ctx, task)
	if err != nil {
		return nil, hperrors.Wrap(err)
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
	scope *entity.ObjectScope,
	requireSettingActive bool,
	requireJobActive bool,
) (backup *entity.Setting, job *entity.Setting, err error) {
	backup, err = uc.SettingRepo.GetSingle(ctx, db, scope, currentSettingType, requireSettingActive)
	if err != nil {
		return nil, nil, hperrors.Wrap(err)
	}

	// Load sched job of the backup
	job, err = uc.SettingRepo.GetSingle(ctx, db, scope, base.SettingTypeSchedJob, requireJobActive,
		bunex.SelectWhere("setting.data->'targetSetting'->>'id' = ?", backup.ID),
	)
	if err != nil {
		return nil, nil, hperrors.Wrap(err)
	}

	return backup, job, nil
}
