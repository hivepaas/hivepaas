package syscleanupserviceimpl

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

var (
	sysCleanupDBModels = []*sysCleanupDBModel{
		{
			Type:  "db/user",
			Model: (*entity.User)(nil),
		},
		{
			Type:  "db/acl-permission",
			Model: (*entity.ACLPermission)(nil),
		},
		{
			Type:         "db/login-trusted-device",
			Model:        (*entity.LoginTrustedDevice)(nil),
			NoSoftDelete: true,
		},
		{
			Type:  "db/setting",
			Model: (*entity.Setting)(nil),
		},
		{
			Type:  "db/project",
			Model: (*entity.Project)(nil),
		},
		{
			Type:  "db/project-tag",
			Model: (*entity.ProjectTag)(nil),
		},
		{
			Type:  "db/project-shared-setting",
			Model: (*entity.ProjectSharedSetting)(nil),
		},
		{
			Type:  "db/app",
			Model: (*entity.App)(nil),
		},
		{
			Type:  "db/app-tag",
			Model: (*entity.AppTag)(nil),
		},
		{
			Type:  "db/deployment",
			Model: (*entity.Deployment)(nil),
		},
		{
			Type:  "db/task",
			Model: (*entity.Task)(nil),
		},
		{
			Type:         "db/task-log",
			Model:        (*entity.TaskLog)(nil),
			NoSoftDelete: true,
		},
		{
			Type:         "db/sys-error",
			Model:        (*entity.SysError)(nil),
			NoSoftDelete: true,
		},
	}
)

type sysCleanupDBModel struct {
	Type         string
	Model        any
	NoSoftDelete bool
}

func (s *service) sysCleanupDB(
	ctx context.Context,
	db database.IDB,
	data *sysCleanupData,
) (err error) {
	retentionSetting := data.SysCleanupSettings.DBObjectRetention
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
	err = s.sysCleanupDBOldDeletedObjects(ctx, db, retentionSetting, timeNow)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Hard delete all old tasks and their logs from the DB
	err = s.sysCleanupDBOldTasks(ctx, db, retentionSetting, timeNow)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Hard delete all old deployments from the DB
	err = s.sysCleanupDBOldDeployments(ctx, db, retentionSetting, timeNow)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Hard delete all old sys-errors from the DB
	err = s.sysCleanupDBOldSysErrors(ctx, db, retentionSetting, timeNow)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}

func (s *service) sysCleanupDBOldDeletedObjects(
	ctx context.Context,
	db database.IDB,
	retentionSetting *entity.DBObjectRetention,
	timeNow time.Time,
) (err error) {
	if retentionSetting.DeletedObjects <= 0 {
		return nil
	}
	oldestTs := timeNow.Add(-retentionSetting.DeletedObjects.ToDuration())
	var errs []error
	for _, model := range sysCleanupDBModels {
		if model.NoSoftDelete {
			continue
		}
		q := db.NewDelete().Model(model.Model).
			ForceDelete().
			WhereAllWithDeleted().
			Where("deleted_at IS NOT NULL").
			Where("deleted_at < ?", oldestTs)
		_, e := q.Exec(ctx)
		if e != nil {
			errs = append(errs, e)
		}
	}
	err = errors.Join(errs...)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (s *service) sysCleanupDBOldTasks(
	ctx context.Context,
	db database.IDB,
	retentionSetting *entity.DBObjectRetention,
	timeNow time.Time,
) (err error) {
	if retentionSetting.Tasks <= 0 {
		return nil
	}

	oldestTs := timeNow.Add(-retentionSetting.Tasks.ToDuration())

	err = s.taskLogRepo.DeleteHard(ctx, db,
		bunex.DeleteWhere("EXISTS(SELECT 1 FROM tasks WHERE tasks.id = task_log.task_id AND "+
			"tasks.updated_at < ?)", oldestTs),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}

	err = s.taskRepo.DeleteHard(ctx, db,
		bunex.DeleteWhere("updated_at < ?", oldestTs),
		bunex.DeleteWithDeleted(),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}

func (s *service) sysCleanupDBOldDeployments(
	ctx context.Context,
	db database.IDB,
	retentionSetting *entity.DBObjectRetention,
	timeNow time.Time,
) (err error) {
	if retentionSetting.Deployments <= 0 {
		return nil
	}

	oldestTs := timeNow.Add(-retentionSetting.Deployments.ToDuration())

	err = s.deploymentRepo.DeleteHard(ctx, db,
		bunex.DeleteWhere("updated_at < ?", oldestTs),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}

func (s *service) sysCleanupDBOldSysErrors(
	ctx context.Context,
	db database.IDB,
	retentionSetting *entity.DBObjectRetention,
	timeNow time.Time,
) (err error) {
	if retentionSetting.SysErrors <= 0 {
		return nil
	}

	oldestTs := timeNow.Add(-retentionSetting.SysErrors.ToDuration())

	err = s.sysErrorRepo.DeleteHard(ctx, db,
		bunex.DeleteWhere("created_at < ?", oldestTs),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}
