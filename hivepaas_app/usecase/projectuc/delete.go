package projectuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
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
			return apperrors.Wrap(err)
		}

		// Remove project and its envs/apps in infra
		err = uc.projectService.DeleteProject(ctx, db, project)
		if err != nil {
			return apperrors.Wrap(err)
		}

		return nil
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &projectdto.DeleteProjectResp{}, nil
}
