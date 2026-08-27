package appserviceimpl

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
)

func (s *service) ExecuteInTx(
	ctx context.Context,
	app *entity.App,
	requireUpdateVerMatch bool,
	fn func(database.Tx) error,
) error {
	err := transaction.Execute(ctx, s.db, func(db database.Tx) error {
		_, err := s.appRepo.GetByID(ctx, db, app.ProjectID, app.ID,
			bunex.SelectColumns("id"),
			bunex.SelectWhereIf(requireUpdateVerMatch, "app.update_ver = ?", app.UpdateVer))
		if err != nil {
			if requireUpdateVerMatch && errors.Is(err, hperrors.ErrNotFound) {
				return hperrors.Wrap(hperrors.ErrUpdateVerMismatched)
			}
			return hperrors.Wrap(err)
		}
		if err = fn(db); err != nil {
			return hperrors.Wrap(err)
		}
		return nil
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
