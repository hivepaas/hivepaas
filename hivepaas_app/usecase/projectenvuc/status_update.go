package projectenvuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/projectenvuc/projectenvdto"
)

func (uc *UC) UpdateProjectEnvStatus(
	ctx context.Context,
	auth *basedto.Auth,
	req *projectenvdto.UpdateProjectEnvStatusReq,
) (*projectenvdto.UpdateProjectEnvStatusResp, error) {
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		projectEnv, err := uc.projectEnvRepo.GetByID(ctx, db, req.ProjectID, req.ProjectEnvID,
			bunex.SelectFor("UPDATE OF project_env"),
			bunex.SelectRelation("Apps",
				bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
			),
		)
		if err != nil {
			return apperrors.Wrap(err)
		}
		if projectEnv.UpdateVer != req.UpdateVer {
			return apperrors.Wrap(apperrors.ErrUpdateVerMismatched)
		}
		// No change
		if projectEnv.Status == req.Status {
			return nil
		}

		err = uc.projectService.SetProjectEnvStatus(ctx, db, projectEnv, req.Status, true)
		if err != nil {
			return apperrors.Wrap(err)
		}
		return nil
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &projectenvdto.UpdateProjectEnvStatusResp{}, nil
}
