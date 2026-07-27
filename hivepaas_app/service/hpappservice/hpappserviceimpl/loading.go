package hpappserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
)

func (s *service) LoadAppByKey(
	ctx context.Context,
	db database.IDB,
	appKey string,
	extraOpts ...bunex.SelectQueryOption,
) (*entity.App, error) {
	opts := []bunex.SelectQueryOption{
		bunex.SelectJoin("JOIN projects ON projects.id = app.project_id"),
		bunex.SelectWhere("projects.key = ?", base.HivepaasProjectKey),
		bunex.SelectWhere("app.key = ?", appKey),
		bunex.SelectLimit(1),
	}
	opts = append(opts, extraOpts...)
	apps, _, err := s.appRepo.List(ctx, db, "", nil, opts...)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	if len(apps) == 0 {
		return nil, apperrors.NewNotFound("App")
	}
	return apps[0], nil
}
