package projectenvuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/auditdetail"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/projectenvuc/projectenvdto"
)

func (uc *UC) DeleteProjectEnv(
	ctx context.Context,
	auth *basedto.Auth,
	req *projectenvdto.DeleteProjectEnvReq,
) (*projectenvdto.DeleteProjectEnvResp, error) {
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		projectEnv, err := uc.projectEnvRepo.GetByID(ctx, db, req.ProjectID, req.ProjectEnvID,
			bunex.SelectFor("UPDATE OF project_env"),
			bunex.SelectRelation("Apps",
				bunex.SelectWhere("app.deleted_at IS NULL"),
			),
		)
		if err != nil {
			return hperrors.Wrap(err)
		}

		// Recorded before the removal, not after. DeleteProjectEnv reaches past the
		// database into the infra, and rolling the transaction back does not put
		// the apps it tore down back; writing the entry first is the only order
		// that cannot leave the removal unrecorded. The app count goes in because
		// it is the size of what was destroyed, and it is unrecoverable after.
		err = uc.recordProjectEnvAction(ctx, db, auth, projectEnv,
			base.AuditLogTypeProjectEnvDelete, base.AuditLogSourceAPIDelete, "env-delete",
			auditdetail.New().
				Set("status", projectEnv.Status).
				Set("appCount", len(projectEnv.Apps)))
		if err != nil {
			return hperrors.Wrap(err)
		}

		// Remove project env and its apps in infra
		err = uc.projectService.DeleteProjectEnv(ctx, db, projectEnv)
		if err != nil {
			return hperrors.Wrap(err)
		}

		return nil
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &projectenvdto.DeleteProjectEnvResp{}, nil
}
