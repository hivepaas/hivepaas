package appserviceimpl

import (
	"context"
	"errors"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/transaction"
)

func (s *service) ExecuteInTx(
	ctx context.Context,
	app *entity.App,
	requireUpdateVerMatch bool,
	fn func(database.Tx) error,
) error {
	err := transaction.Execute(ctx, s.db, func(db database.Tx) error {
		_, err := s.appRepo.GetByID(ctx, db, "", app.ID,
			bunex.SelectColumns("id"),
			bunex.SelectWhereIf(requireUpdateVerMatch, "app.update_ver = ?", app.UpdateVer))
		if err != nil {
			if requireUpdateVerMatch && errors.Is(err, apperrors.ErrNotFound) {
				return apperrors.New(apperrors.ErrUpdateVerMismatched)
			}
			return apperrors.New(err)
		}
		if err = fn(db); err != nil {
			return apperrors.New(err)
		}
		return nil
	})
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}
