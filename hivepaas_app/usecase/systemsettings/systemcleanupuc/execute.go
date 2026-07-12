package systemcleanupuc

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
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/systemsettings/systemcleanupuc/systemcleanupdto"
)

func (uc *UC) ExecuteSystemCleanup(
	ctx context.Context,
	auth *basedto.Auth,
	req *systemcleanupdto.ExecuteSystemCleanupReq,
) (*systemcleanupdto.ExecuteSystemCleanupResp, error) {
	req.Type = currentSettingType
	_, jobSetting, err := uc.getCleanupSettingAndJob(ctx, uc.DB, req.Scope, true, false)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	task, err := uc.schedJobService.CreateSchedJobTask(jobSetting, time.Time{}, timeutil.NowUTC())
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

	return &systemcleanupdto.ExecuteSystemCleanupResp{
		Data: &systemcleanupdto.ExecuteSystemCleanupDataResp{
			Task: &basedto.ObjectIDResp{ID: task.ID},
		},
	}, nil
}

func (uc *UC) getCleanupSettingAndJob(
	ctx context.Context,
	db database.IDB,
	scope *base.ObjectScope,
	requireSettingActive bool,
	requireJobActive bool,
) (cleanup *entity.Setting, job *entity.Setting, err error) {
	cleanup, err = uc.SettingRepo.GetSingle(ctx, db, scope, currentSettingType, requireSettingActive)
	if err != nil {
		return nil, nil, apperrors.Wrap(err)
	}

	// Load sched job of the cleanup
	job, err = uc.SettingRepo.GetSingle(ctx, db, scope, base.SettingTypeSchedJob, requireJobActive,
		bunex.SelectWhere("setting.data->'targetSetting'->>'id' = ?", cleanup.ID),
	)
	if err != nil {
		return nil, nil, apperrors.Wrap(err)
	}

	return cleanup, job, nil
}
