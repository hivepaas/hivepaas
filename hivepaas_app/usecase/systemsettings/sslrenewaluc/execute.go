package sslrenewaluc

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
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/systemsettings/sslrenewaluc/sslrenewaldto"
)

func (uc *UC) ExecuteSSLRenewal(
	ctx context.Context,
	auth *basedto.Auth,
	req *sslrenewaldto.ExecuteSSLRenewalReq,
) (*sslrenewaldto.ExecuteSSLRenewalResp, error) {
	req.Type = currentSettingType
	_, jobSetting, err := uc.getRenewalSettingAndJob(ctx, uc.DB, req.Scope, true, false)
	if err != nil {
		return nil, apperrors.New(err)
	}

	task, err := uc.schedJobService.CreateSchedJobTask(jobSetting, time.Time{}, timeutil.NowUTC())
	if err != nil {
		return nil, apperrors.New(err)
	}

	// If no specific settings to be sent, we will try to renew all renewable SSLs
	if len(req.TargetSSLs) > 0 {
		task.MustSetArgs(&entity.TaskSSLRenewalArgs{
			TargetSSLs: req.TargetSSLs.ToEntity(),
		})
	}

	err = uc.taskRepo.Insert(ctx, uc.DB, task)
	if err != nil {
		return nil, apperrors.New(err)
	}

	err = uc.taskQueue.ScheduleTask(ctx, task)
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &sslrenewaldto.ExecuteSSLRenewalResp{
		Data: &sslrenewaldto.ExecuteSSLRenewalDataResp{
			Task: &basedto.ObjectIDResp{ID: task.ID},
		},
	}, nil
}

func (uc *UC) getRenewalSettingAndJob(
	ctx context.Context,
	db database.IDB,
	scope *base.ObjectScope,
	requireSettingActive bool,
	requireJobActive bool,
) (cleanup *entity.Setting, job *entity.Setting, err error) {
	cleanup, err = uc.SettingRepo.GetSingle(ctx, db, scope, currentSettingType, requireSettingActive)
	if err != nil {
		return nil, nil, apperrors.New(err)
	}

	// Load sched job of the renewal
	job, err = uc.SettingRepo.GetSingle(ctx, db, scope, base.SettingTypeSchedJob, requireJobActive,
		bunex.SelectWhere("setting.data->'targetSetting'->>'id' = ?", cleanup.ID),
	)
	if err != nil {
		return nil, nil, apperrors.New(err)
	}

	return cleanup, job, nil
}
