package syscleanupserviceimpl

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/auditdetail"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
)

// orphanedAppCondition finds the apps nothing owns: their project, or their
// env, is gone. A project being deleted is not gone - its row stays until its
// envs are - so an app a deletion is still working through is not taken here.
const orphanedAppCondition = "NOT EXISTS (SELECT 1 FROM projects" +
	" WHERE projects.id = app.project_id AND projects.deleted_at IS NULL)" +
	" OR NOT EXISTS (SELECT 1 FROM project_envs" +
	" WHERE project_envs.id = app.project_env_id AND project_envs.deleted_at IS NULL)"

// sysCleanupDBDeleteOrphanedApps deletes the apps whose project or env is gone.
//
// Deleting a project deletes its envs one by one and goes on when one fails, so
// an env that could not be deleted leaves its apps behind a project that no
// longer exists: their services still run, their domains are still taken, and
// no screen can reach them to remove them. They are deleted as a person would
// delete them - service, settings, tasks, routing - with their data kept: which
// volumes were meant to go with the project is not known any more.
//
// Each app is deleted in a savepoint of its own, so one that fails is rolled
// back alone and the rest go on; one being worked on elsewhere is skipped, and
// found again tomorrow.
func (s *service) sysCleanupDBDeleteOrphanedApps(
	ctx context.Context,
	db database.IDB,
	data *sysCleanupData,
) (err error) {
	orphans, _, err := s.appRepo.List(ctx, db, "", nil,
		bunex.SelectColumns("app.id", "app.parent_id"),
		bunex.SelectWhere(orphanedAppCondition),
		// A parent deletes its children: they come after it, and are found gone.
		bunex.SelectOrder("app.parent_id ASC NULLS FIRST", "app.id"),
	)
	if err != nil {
		return hperrors.Wrap(err)
	}

	for _, orphan := range orphans {
		out, e := s.deleteOrphanedApp(ctx, db, orphan.ID)
		if out == nil {
			err = errors.Join(err, e)
			continue
		}
		if e != nil {
			out.Error = e.Error()
			err = errors.Join(err, e)
			_ = data.LogStore.Add(ctx, tasklog.NewErrFrame("Failed to delete orphaned app "+out.Name+
				" ("+out.ID+"): "+e.Error(), tasklog.TsNow))
		} else {
			_ = data.LogStore.Add(ctx, tasklog.NewOutFrame("Orphaned app deleted: "+out.Name+
				" ("+out.ID+")", tasklog.TsNow))
		}
		data.TaskOutput.DBCleanup.OrphanedAppsDeleted = append(data.TaskOutput.DBCleanup.OrphanedAppsDeleted, out)
	}
	return hperrors.Wrap(err)
}

// deleteOrphanedApp deletes one orphaned app in a savepoint. It answers nil
// output for an app it did not try: deleted already, as a child of another, or
// locked by whatever is working on it.
func (s *service) deleteOrphanedApp(ctx context.Context, db database.IDB, appID string) (
	out *entity.OrphanedAppOutput, err error) {
	err = transaction.Execute(ctx, db, func(tx database.Tx) error {
		app, e := s.appRepo.GetByID(ctx, tx, "", appID,
			bunex.SelectWhere(orphanedAppCondition),
			bunex.SelectFor("UPDATE OF app SKIP LOCKED"),
		)
		if e != nil {
			if errors.Is(e, hperrors.ErrNotFound) {
				return nil
			}
			return hperrors.Wrap(e)
		}
		out = &entity.OrphanedAppOutput{ID: app.ID, Name: app.Name}

		// Recorded first, as a person's delete is: the removal reaches past the
		// database, and a rollback does not put back what it tore down.
		e = s.auditService.Record(ctx, tx, &auditservice.Entry{
			Type:     base.AuditLogTypeAppDelete,
			Scope:    base.ObjectScopeApp,
			ObjectID: app.ID,
			Source:   base.AuditLogSourceSystemCleanup,
			Result:   base.AuditLogResultAllowed,
			ResType:  base.ResourceTypeApp,
			ResID:    app.ID,
			ResName:  app.Name,
			Detail: auditdetail.New().
				Set("reason", "orphaned: its project or env no longer exists").
				Set("projectId", app.ProjectID).
				Set("envId", app.ProjectEnvID).
				Set("serviceId", app.ServiceID).
				Set("removeStorage", false).String(),
		})
		if e != nil {
			return hperrors.Wrap(e)
		}
		return hperrors.Wrap(s.appService.DeleteApp(ctx, tx, app, false, true))
	})
	return out, hperrors.Wrap(err)
}
