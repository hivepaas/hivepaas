package projectenvuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
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
			return apperrors.Wrap(err)
		}

		// Remove project env and its apps in infra
		err = uc.projectService.DeleteProjectEnv(ctx, db, projectEnv)
		if err != nil {
			return apperrors.Wrap(err)
		}

		return nil
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &projectenvdto.DeleteProjectEnvResp{}, nil
}
