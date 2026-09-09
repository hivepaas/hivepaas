package syscleanupserviceimpl

import (
	"context"
	"errors"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/entityutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
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
			Type:  "db/res_link",
			Model: (*entity.ResLink)(nil),
		},
		{
			Type:  "db/tag",
			Model: (*entity.Tag)(nil),
		},
		{
			Type:  "db/project",
			Model: (*entity.Project)(nil),
		},
		{
			Type:  "db/project-shared-setting",
			Model: (*entity.SharedSetting)(nil),
		},
		{
			Type:  "db/app",
			Model: (*entity.App)(nil),
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
		{
			Type:         "db/audit-log",
			Model:        (*entity.AuditLog)(nil),
			NoSoftDelete: true,
		},
		{
			Type:  "db/bin-object",
			Model: (*entity.BinObject)(nil),
		},
		{
			Type:  "db/file",
			Model: (*entity.File)(nil),
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
	retentionSetting := &data.SysCleanupSettings.DBObjectRetention
	if !retentionSetting.Enabled {
		return nil
	}

	defer func() {
		if err != nil {
			data.TaskOutput.DBCleanup.Error = err.Error()
		}
	}()

	timeNow := timeutil.NowUTC()

	// Soft delete all orphaned tasks and deployments belonging to deleted apps
	e := s.sysCleanupDBDeleteOrphanedTasksAndDeployments(ctx, db)
	if e != nil {
		err = errors.Join(err, e)
	}

	// Hard delete all old deleted objects from the DB
	e = s.sysCleanupDBOldDeletedObjects(ctx, db, retentionSetting, timeNow)
	if e != nil {
		err = errors.Join(err, e)
	}

	// Hard delete all old tasks and their logs from the DB
	e = s.sysCleanupDBOldTasks(ctx, db, retentionSetting, timeNow)
	if e != nil {
		err = errors.Join(err, e)
	}

	// Hard delete all old deployments from the DB
	e = s.sysCleanupDBOldDeployments(ctx, db, retentionSetting, timeNow)
	if e != nil {
		err = errors.Join(err, e)
	}

	// Hard delete all old sys-errors from the DB
	e = s.sysCleanupDBOldSysErrors(ctx, db, retentionSetting, timeNow)
	if e != nil {
		err = errors.Join(err, e)
	}

	// Hard delete all old audit logs from the DB
	e = s.sysCleanupDBOldAuditLogs(ctx, db, retentionSetting, timeNow)
	if e != nil {
		err = errors.Join(err, e)
	}

	// Hard delete all old locks from the DB
	e = s.sysCleanupDBOldLocks(ctx, db, timeNow)
	if e != nil {
		err = errors.Join(err, e)
	}

	return nil
}

func (s *service) sysCleanupDBDeleteOrphanedTasksAndDeployments(
	ctx context.Context,
	db database.IDB,
) (err error) {
	// Soft delete tasks belonging to deleted apps (skipping currently locked ones)
	orphanedTasks, _, e := s.taskRepo.ListByTarget(ctx, db, "", nil,
		bunex.SelectColumns("id"),
		bunex.SelectWhere("EXISTS(SELECT 1 FROM apps "+
			"WHERE apps.id = task.target_id AND apps.deleted_at IS NOT NULL) OR "+
			"EXISTS(SELECT 1 FROM deployments JOIN apps ON apps.id = deployments.app_id "+
			"WHERE deployments.id = task.target_id AND apps.deleted_at IS NOT NULL)"),
		bunex.SelectFor("UPDATE SKIP LOCKED"),
	)
	if e != nil {
		err = errors.Join(err, e)
	}
	e = s.taskRepo.DeleteByIDs(ctx, db, entityutil.ExtractIDs(orphanedTasks))
	if e != nil {
		err = errors.Join(err, e)
	}

	// Soft delete deployments belonging to deleted apps (skipping currently locked ones)
	orphanedDeployments, _, e := s.deploymentRepo.List(ctx, db, "", nil,
		bunex.SelectColumns("id"),
		bunex.SelectWhere("EXISTS(SELECT 1 FROM apps "+
			"WHERE apps.id = deployment.app_id AND apps.deleted_at IS NOT NULL)"),
		bunex.SelectFor("UPDATE SKIP LOCKED"),
	)
	if e != nil {
		err = errors.Join(err, e)
	}
	e = s.deploymentRepo.DeleteByIDs(ctx, db, entityutil.ExtractIDs(orphanedDeployments))
	if e != nil {
		err = errors.Join(err, e)
	}

	return hperrors.Wrap(err)
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
		return hperrors.Wrap(err)
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
		bunex.DeleteWhere("ts < ?", oldestTs),
		// bunex.DeleteWhere("EXISTS(SELECT 1 FROM tasks WHERE tasks.id = task_log.task_id AND "+
		//	"tasks.updated_at < ?)", oldestTs),
	)
	if err != nil {
		return hperrors.Wrap(err)
	}

	err = s.taskRepo.DeleteHard(ctx, db,
		bunex.DeleteWhere("updated_at < ?", oldestTs),
	)
	if err != nil {
		return hperrors.Wrap(err)
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
		return hperrors.Wrap(err)
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
		return hperrors.Wrap(err)
	}

	return nil
}

// auditLogRevealRetention is the shortest time a secret-reveal entry is kept,
// whatever the configured retention says.
//
// Retention is an operator setting, and without a floor, shortening it would be a
// supported way to erase the record of one's own reveals - the one thing this log
// exists to make impossible. A longer configured retention still wins: the floor
// only stops entries being aged out early, it does not cap how long they are kept.
const auditLogRevealRetention = 90 * timeutil.Day

// auditLogCutoffs returns the instant before which an entry may be deleted: one
// for entries in general, one for the reveal entries that carry the floor.
//
// Both are instants, so the earlier of the two is the longer retention - which is
// why a configured retention longer than the floor is kept as it is.
func auditLogCutoffs(retention time.Duration, timeNow time.Time) (oldestTs, revealOldestTs time.Time) {
	oldestTs = timeNow.Add(-retention)
	revealOldestTs = timeNow.Add(-auditLogRevealRetention)
	if oldestTs.Before(revealOldestTs) {
		revealOldestTs = oldestTs
	}
	return oldestTs, revealOldestTs
}

func (s *service) sysCleanupDBOldAuditLogs(
	ctx context.Context,
	db database.IDB,
	retentionSetting *entity.DBObjectRetention,
	timeNow time.Time,
) (err error) {
	if retentionSetting.AuditLogs <= 0 {
		return nil
	}

	oldestTs, revealOldestTs := auditLogCutoffs(retentionSetting.AuditLogs.ToDuration(), timeNow)

	err = s.auditLogRepo.DeleteHard(ctx, db,
		bunex.DeleteWhere("CASE WHEN type = ? THEN created_at < ? ELSE created_at < ? END",
			base.AuditLogTypeSecretReveal, revealOldestTs, oldestTs),
	)
	if err != nil {
		return hperrors.Wrap(err)
	}

	return nil
}

func (s *service) sysCleanupDBOldLocks(
	ctx context.Context,
	db database.IDB,
	timeNow time.Time,
) (err error) {
	oldestTs := timeNow.Add(-90 * 24 * time.Hour) //nolint:mnd

	// Find all old locks with skipping locked ones
	deletingLocks, _, err := s.lockRepo.List(ctx, db, nil,
		bunex.SelectWhere("created_at < ?", oldestTs),
		bunex.SelectFor("UPDATE SKIP LOCKED"),
	)
	if err != nil {
		return hperrors.Wrap(err)
	}

	err = s.lockRepo.DeleteByIDs(ctx, db, entityutil.ExtractIDs(deletingLocks))
	if err != nil {
		return hperrors.Wrap(err)
	}

	return nil
}
