package appuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
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
		)
		if err != nil {
			return hperrors.Wrap(err)
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
				Set("serviceId", app.ServiceID))
		if err != nil {
			return hperrors.Wrap(err)
		}

		// Remove app and its data from the infra
		err = uc.appService.DeleteApp(ctx, db, app)
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
