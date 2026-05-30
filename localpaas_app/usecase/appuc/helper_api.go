package appuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
)

func (uc *UC) GetProjectIDForAppKey(
	ctx context.Context,
	appKey string,
	requireActive bool,
) (string, error) {
	app, err := uc.appRepo.GetByKey(ctx, uc.db, "", appKey,
		bunex.SelectColumns("project_id"),
		bunex.SelectWhereIf(requireActive, "app.status = ?", base.AppStatusActive),
	)
	if err != nil {
		return "", apperrors.Wrap(err)
	}
	return app.ProjectID, nil
}
