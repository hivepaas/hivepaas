package projectserviceimpl

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
)

func (s *service) ExecuteEnvInTx(
	ctx context.Context,
	projectEnv *entity.ProjectEnv,
	requireUpdateVerMatch bool,
	fn func(database.Tx) error,
) error {
	err := transaction.Execute(ctx, s.db, func(db database.Tx) error {
		_, err := s.projectEnvRepo.GetByID(ctx, db, projectEnv.ProjectID, projectEnv.ID,
			bunex.SelectColumns("id"),
			bunex.SelectWhereIf(requireUpdateVerMatch, "project_env.update_ver = ?", projectEnv.UpdateVer))
		if err != nil {
			if requireUpdateVerMatch && errors.Is(err, apperrors.ErrNotFound) {
				return apperrors.Wrap(apperrors.ErrUpdateVerMismatched)
			}
			return apperrors.Wrap(err)
		}
		if err = fn(db); err != nil {
			return apperrors.Wrap(err)
		}
		return nil
	})
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}
