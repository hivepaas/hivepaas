package appuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
)

func (uc *UC) GetProjectIDForAppKey(
	ctx context.Context,
	appKey string,
	requireActive bool,
) (string, error) {
	app, err := uc.appRepo.GetByGlobalKey(ctx, uc.db, "", appKey,
		bunex.SelectColumns("project_id"),
		bunex.SelectWhereIf(requireActive, "app.status = ?", base.AppStatusActive),
	)
	if err != nil {
		return "", apperrors.Wrap(err)
	}
	return app.ProjectID, nil
}
