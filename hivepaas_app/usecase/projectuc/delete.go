package projectuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/auditdetail"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/projectuc/projectdto"
)

func (uc *UC) DeleteProject(
	ctx context.Context,
	auth *basedto.Auth,
	req *projectdto.DeleteProjectReq,
) (*projectdto.DeleteProjectResp, error) {
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		project, err := uc.projectRepo.GetByID(ctx, db, req.ProjectID,
			bunex.SelectFor("UPDATE OF project"),
			bunex.SelectRelation("ProjectEnvs.Apps"),
		)
		if err != nil {
			return hperrors.Wrap(err)
		}

		// Recorded before the removal. DeleteProject reaches past the database into
		// the infra, and rolling the transaction back does not put back the
		// environments and apps it tore down; writing the entry first is the only
		// order that cannot leave the removal unrecorded.
		//
		// The counts go in because they are the size of what was destroyed, and
		// after this there is nothing left to count.
		appCount := 0
		for _, projectEnv := range project.ProjectEnvs {
			appCount += len(projectEnv.Apps)
		}
		err = uc.recordProjectWrite(ctx, db, auth, project,
			base.AuditLogTypeProjectDelete, base.AuditLogSourceAPIDelete, "delete", auditdetail.New().
				Set("key", project.Key).
				Set("ownerId", project.OwnerID).
				Set("status", project.Status).
				Set("envCount", len(project.ProjectEnvs)).
				Set("appCount", appCount))
		if err != nil {
			return hperrors.Wrap(err)
		}

		// Remove project and its envs/apps in infra
		err = uc.projectService.DeleteProject(ctx, db, project)
		if err != nil {
			return hperrors.Wrap(err)
		}

		return nil
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &projectdto.DeleteProjectResp{}, nil
}
