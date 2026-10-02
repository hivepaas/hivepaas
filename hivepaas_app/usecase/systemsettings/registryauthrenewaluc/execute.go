package registryauthrenewaluc

import (
	"context"
	"errors"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/systemsettings/registryauthrenewaluc/registryauthrenewaldto"
)

func (uc *UC) ExecuteRegistryAuthRenewal(
	ctx context.Context,
	auth *basedto.Auth,
	req *registryauthrenewaldto.ExecuteRegistryAuthRenewalReq,
) (*registryauthrenewaldto.ExecuteRegistryAuthRenewalResp, error) {
	req.Type = currentSettingType
	task, err := uc.scheduleRenewal(ctx, uc.DB, req.TargetAuths.ToEntity(), true)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &registryauthrenewaldto.ExecuteRegistryAuthRenewalResp{
		Data: &registryauthrenewaldto.ExecuteRegistryAuthRenewalDataResp{
			Task: &basedto.ObjectIDResp{ID: task.ID},
		},
	}, nil
}

// RenewOnSave runs the renewal at once for a credential whose AWS keys were
// saved: its services get a token got with the new keys. A renewal turned off,
// or not made yet, is not run. The task is inserted with db, the caller's
// transaction, and scheduled by the function returned, to call once that has
// committed.
func (uc *UC) RenewOnSave(ctx context.Context, db database.IDB, authID string) (schedule func(), err error) {
	jobSetting, err := uc.activeJob(ctx, db)
	if err != nil || jobSetting == nil {
		return func() {}, err
	}
	task, err := uc.insertRenewalTask(ctx, db, jobSetting, entity.ObjectIDSlice{{ID: authID}})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return func() { _ = uc.taskQueue.ScheduleTask(context.WithoutCancel(ctx), task) }, nil
}

// RenewIfStale runs the renewal at start when its last run ended longer ago than
// its interval, and an ECR credential is there to renew: HivePaaS down for a
// while may have let the tokens Swarm keeps age past what the schedule allows.
func (uc *UC) RenewIfStale(ctx context.Context) (bool, error) {
	jobSetting, err := uc.activeJob(ctx, uc.DB)
	if err != nil || jobSetting == nil {
		return false, err
	}
	ecrAuths, _, err := uc.SettingRepo.List(ctx, uc.DB, nil, nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeRegistryAuth),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
		bunex.SelectWhere("setting.data->>'kind' = ?", base.RegistryAuthKindAWSECR),
		bunex.SelectColumns("id"),
		bunex.SelectLimit(1),
	)
	if err != nil {
		return false, hperrors.Wrap(err)
	}
	if len(ecrAuths) == 0 {
		return false, nil
	}

	lastRuns, _, err := uc.taskRepo.ListByTarget(ctx, uc.DB, jobSetting.ID, nil,
		bunex.SelectWhere("task.status = ?", base.TaskStatusDone),
		bunex.SelectOrder("task.ended_at DESC"),
		bunex.SelectLimit(1),
	)
	if err != nil {
		return false, hperrors.Wrap(err)
	}
	if len(lastRuns) > 0 && timeutil.NowUTC().Sub(lastRuns[0].EndedAt) < jobInterval(jobSetting) {
		return false, nil
	}

	if _, err = uc.scheduleRenewal(ctx, uc.DB, nil, false); err != nil {
		return false, hperrors.Wrap(err)
	}
	return true, nil
}

// scheduleRenewal inserts a renewal task for the targets, every ECR credential
// when none, and schedules it.
func (uc *UC) scheduleRenewal(
	ctx context.Context,
	db database.IDB,
	targets entity.ObjectIDSlice,
	requireSettingActive bool,
) (*entity.Task, error) {
	_, jobSetting, err := uc.getRenewalSettingAndJob(ctx, db, requireSettingActive)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	task, err := uc.insertRenewalTask(ctx, db, jobSetting, targets)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if err = uc.taskQueue.ScheduleTask(ctx, task); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return task, nil
}

func (uc *UC) insertRenewalTask(
	ctx context.Context,
	db database.IDB,
	jobSetting *entity.Setting,
	targets entity.ObjectIDSlice,
) (*entity.Task, error) {
	timeNow := timeutil.NowUTC()
	// A task's run time is unique for its job: one a scheduled run already
	// has is moved by a second.
	task, err := uc.schedJobService.CreateSchedJobTask(jobSetting, timeNow.Add(time.Second), timeNow)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if len(targets) > 0 {
		task.MustSetArgs(&entity.TaskRegistryAuthRenewalArgs{TargetAuths: targets})
	}
	if err = uc.taskRepo.Insert(ctx, db, task); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return task, nil
}

// activeJob is the renewal's job when it and its setting are on; nil when either
// is off or not made yet.
func (uc *UC) activeJob(ctx context.Context, db database.IDB) (*entity.Setting, error) {
	renewal, job, err := uc.getRenewalSettingAndJob(ctx, db, false)
	if errors.Is(err, hperrors.ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if renewal.Status != base.SettingStatusActive || job.Status != base.SettingStatusActive {
		return nil, nil
	}
	return job, nil
}

// jobInterval is the interval of the job's schedule, as the renewal allows it.
func jobInterval(job *entity.Setting) time.Duration {
	schedJob, err := job.AsSchedJob()
	if err != nil || schedJob.Schedule == nil {
		return entity.RegistryAuthRenewalIntervalDefault
	}
	return (&entity.RegistryAuthRenewal{Schedule: *schedJob.Schedule}).Interval()
}

func (uc *UC) getRenewalSettingAndJob(
	ctx context.Context,
	db database.IDB,
	requireSettingActive bool,
) (renewal *entity.Setting, job *entity.Setting, err error) {
	scope := entity.NewObjectScopeGlobal()
	renewal, err = uc.SettingRepo.GetSingle(ctx, db, scope, currentSettingType, requireSettingActive)
	if err != nil {
		return nil, nil, hperrors.Wrap(err)
	}

	job, err = uc.SettingRepo.GetSingle(ctx, db, scope, base.SettingTypeSchedJob, false,
		bunex.SelectWhere("setting.kind = ?", base.SchedJobTypeRegistryAuthRenewal),
		bunex.SelectWhere("setting.data->'targetSetting'->>'id' = ?", renewal.ID),
	)
	if err != nil {
		return nil, nil, hperrors.Wrap(err)
	}

	return renewal, job, nil
}
