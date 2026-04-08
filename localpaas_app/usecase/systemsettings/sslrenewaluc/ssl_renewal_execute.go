package sslrenewaluc

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
	"github.com/localpaas/localpaas/localpaas_app/usecase/systemsettings/sslrenewaluc/sslrenewaldto"
)

func (uc *UC) ExecuteSSLRenewal(
	ctx context.Context,
	auth *basedto.Auth,
	req *sslrenewaldto.ExecuteSSLRenewalReq,
) (*sslrenewaldto.ExecuteSSLRenewalResp, error) {
	req.Type = currentSettingType
	_, jobSetting, err := uc.getRenewalSettingAndJob(ctx, uc.DB, req.Scope, true, false)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	task, err := uc.cronJobService.CreateCronJobTask(jobSetting, time.Time{}, timeutil.NowUTC())
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	// If no specific settings to be sent, we will try to renew all renewable SSLs
	if len(req.TargetSSLs) > 0 {
		task.MustSetArgs(&entity.TaskSSLRenewalArgs{
			TargetSSLs: req.TargetSSLs.ToEntity(),
		})
	}

	err = uc.taskRepo.Insert(ctx, uc.DB, task)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	err = uc.taskQueue.ScheduleTask(ctx, task)
	if err != nil {
		return nil, apperrors.Wrap(err)
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
	scope *base.SettingScope,
	requireSettingActive bool,
	requireJobActive bool,
) (cleanup *entity.Setting, job *entity.Setting, err error) {
	cleanup, err = uc.SettingRepo.GetSingle(ctx, db, scope, currentSettingType, requireSettingActive)
	if err != nil {
		return nil, nil, apperrors.Wrap(err)
	}

	// Load cron job of the renewal
	job, err = uc.SettingRepo.GetSingle(ctx, db, scope, base.SettingTypeCronJob, requireJobActive,
		bunex.SelectWhere("setting.data->'targetSetting'->>'id' = ?", cleanup.ID),
	)
	if err != nil {
		return nil, nil, apperrors.Wrap(err)
	}

	return cleanup, job, nil
}
