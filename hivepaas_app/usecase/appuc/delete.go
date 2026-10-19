package appuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/auditdetail"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appuc/appdto"
)

func (uc *UC) DeleteApp(
	ctx context.Context,
	auth *basedto.Auth,
	req *appdto.DeleteAppReq,
) (*appdto.DeleteAppResp, error) {
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		app, err := uc.appService.LoadApp(ctx, db, req.ProjectID, req.AppID, false, false,
			bunex.SelectFor("UPDATE OF app"),
			bunex.SelectRelation("Project", bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...)),
			bunex.SelectRelation("Settings",
				bunex.SelectWhere("setting.type = ?", base.SettingTypeAppManaged),
				bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
			),
		)
		if err != nil {
			return hperrors.Wrap(err)
		}
		if err = entity.CheckAppDeletion(app.Project, app); err != nil {
			return hperrors.Wrap(err)
		}
		if err = refuseClusterDeletion(app, req.DeleteCluster); err != nil {
			return err
		}

		// Recorded before the removal. DeleteApp reaches past the database into the
		// infra, and rolling the transaction back does not put back what it tore
		// down; writing the entry first is the only order that cannot leave the
		// removal unrecorded. The identifiers go in because after this they are
		// gone, and an entry naming an app that no longer exists is the entry
		// somebody will be reading.
		err = uc.recordAppWrite(ctx, db, auth, app,
			base.AuditLogTypeAppDelete, base.AuditLogSourceAPIDelete, "delete", auditdetail.New().
				Set("projectId", app.ProjectID).
				Set("envId", app.ProjectEnvID).
				Set("status", app.Status).
				Set("serviceId", app.ServiceID).
				Set("removeStorage", req.RemoveStorage).
				Set("deleteCluster", req.DeleteCluster))
		if err != nil {
			return hperrors.Wrap(err)
		}

		// Remove app and its data from the infra
		err = uc.appService.DeleteApp(ctx, db, app, req.RemoveStorage, true)
		if err != nil {
			return hperrors.Wrap(err)
		}

		return nil
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &appdto.DeleteAppResp{}, nil
}

// refuseClusterDeletion keeps a managed cluster whole: a member is removed from
// its primary's Database tab, which keeps the members it lists true; and the
// app a replica was promoted over still holds the cluster - its setting, its
// jobs and its members - so it is deleted only when the whole cluster is asked
// for: deleting it deletes the primary with it.
func refuseClusterDeletion(app *entity.App, deleteCluster bool) error {
	setting := app.GetSettingByType(base.SettingTypeAppManaged)
	if setting == nil {
		return nil
	}
	managed, err := setting.AsAppManagedSettings()
	if err != nil {
		return hperrors.Wrap(err)
	}
	if managed.MemberOf != nil {
		return hperrors.Wrap(hperrors.ErrAppManagedIsMember).WithParam("App", app.Name).ExposeParams("App")
	}
	if managed.Postgres != nil && managed.Postgres.Writer.ID != "" && !deleteCluster {
		return hperrors.Wrap(hperrors.ErrAppManagedClusterHeld).WithParam("App", app.Name).ExposeParams("App")
	}
	return nil
}
