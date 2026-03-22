package taskcronjobexec

import (
	"context"
	"errors"
	"time"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
)

func (e *Executor) sysDBCleanup(
	ctx context.Context,
	db database.IDB,
	retentionSetting *entity.DBObjectRetention,
	data *sysCleanupTaskData,
) (err error) {
	if retentionSetting == nil || !retentionSetting.Enabled {
		return nil
	}
	timeNow := timeutil.NowUTC()
	defer func() {
		if err != nil {
			data.TaskOutput.DBCleanup.Error = err.Error()
		}
	}()

	// Hard delete all old deleted objects from the DB
	err = e.sysCleanupOldDeletedObjects(ctx, db, retentionSetting, timeNow)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Hard delete all old tasks and their logs from the DB
	err = e.sysCleanupOldTasks(ctx, db, retentionSetting, timeNow)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Hard delete all old deployments from the DB
	err = e.sysCleanupOldDeployments(ctx, db, retentionSetting, timeNow)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Hard delete all old sys-errors from the DB
	err = e.sysCleanupOldSysErrors(ctx, db, retentionSetting, timeNow)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}

func (e *Executor) sysCleanupOldDeletedObjects(
	ctx context.Context,
	db database.IDB,
	retentionSetting *entity.DBObjectRetention,
	timeNow time.Time,
) (err error) {
	if retentionSetting.DeletedObjects <= 0 {
		return nil
	}

	oldestTs := timeNow.Add(-retentionSetting.DeletedObjects.ToDuration())
	opts := []bunex.DeleteQueryOption{
		bunex.DeleteWithDeleted(),
		bunex.DeleteWhere("deleted_at IS NOT NULL"),
		bunex.DeleteWhere("deleted_at < ?", oldestTs),
	}
	// NOTE: apply on tables with soft-deleted enabled
	e1 := e.userRepo.DeleteHard(ctx, db, opts...)
	e2 := e.aclPermissionRepo.DeleteHard(ctx, db, opts...)
	e3 := e.projectRepo.DeleteHard(ctx, db, opts...)
	e4 := e.projectTagRepo.DeleteHard(ctx, db, opts...)
	e5 := e.projectSharedSettingRepo.DeleteHard(ctx, db, opts...)
	e6 := e.appRepo.DeleteHard(ctx, db, opts...)
	e7 := e.appTagRepo.DeleteHard(ctx, db, opts...)
	e8 := e.deploymentRepo.DeleteHard(ctx, db, opts...)
	e9 := e.settingRepo.DeleteHard(ctx, db, opts...)
	e10 := e.taskRepo.DeleteHard(ctx, db, opts...)
	err = errors.Join(e1, e2, e3, e4, e5, e6, e7, e8, e9, e10)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}

func (e *Executor) sysCleanupOldTasks(
	ctx context.Context,
	db database.IDB,
	retentionSetting *entity.DBObjectRetention,
	timeNow time.Time,
) (err error) {
	if retentionSetting.Tasks <= 0 {
		return nil
	}

	oldestTs := timeNow.Add(-retentionSetting.Tasks.ToDuration())

	err = e.taskLogRepo.DeleteHard(ctx, db,
		bunex.DeleteWhere("EXISTS(SELECT 1 FROM tasks WHERE tasks.id = task_log.task_id AND "+
			"tasks.updated_at < ?)", oldestTs),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}

	err = e.taskRepo.DeleteHard(ctx, db,
		bunex.DeleteWhere("updated_at < ?", oldestTs),
		bunex.DeleteWithDeleted(),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}

func (e *Executor) sysCleanupOldDeployments(
	ctx context.Context,
	db database.IDB,
	retentionSetting *entity.DBObjectRetention,
	timeNow time.Time,
) (err error) {
	if retentionSetting.Deployments <= 0 {
		return nil
	}

	oldestTs := timeNow.Add(-retentionSetting.Deployments.ToDuration())

	err = e.deploymentRepo.DeleteHard(ctx, db,
		bunex.DeleteWhere("updated_at < ?", oldestTs),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}

func (e *Executor) sysCleanupOldSysErrors(
	ctx context.Context,
	db database.IDB,
	retentionSetting *entity.DBObjectRetention,
	timeNow time.Time,
) (err error) {
	if retentionSetting.SysErrors <= 0 {
		return nil
	}

	oldestTs := timeNow.Add(-retentionSetting.SysErrors.ToDuration())

	err = e.sysErrorRepo.DeleteHard(ctx, db,
		bunex.DeleteWhere("created_at < ?", oldestTs),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}
