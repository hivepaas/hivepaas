package apppreviewserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
)

func (s *service) GetPreview(
	ctx context.Context,
	db database.IDB,
	appID, repoRef string,
	extraOpts ...bunex.SelectQueryOption,
) (*entity.App, error) {
	listOpts := []bunex.SelectQueryOption{
		bunex.SelectWhere("app.parent_id = ?", appID),
		bunex.SelectRelation("Settings",
			bunex.SelectWhere("setting.type = ?", base.SettingTypeAppDeployment),
			bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
			bunex.SelectWhereIf(repoRef != "", "setting.data->'repoSource'->>'repoRef' = ?", repoRef),
		),
	}
	listOpts = append(listOpts, extraOpts...)

	apps, _, err := s.appRepo.List(ctx, db, "", nil, listOpts...)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	for _, app := range apps {
		if app.GetSettingByType(base.SettingTypeAppDeployment) != nil {
			return app, nil
		}
	}
	return nil, apperrors.NewNotFound("App")
}

func (s *service) GetPreviews(
	ctx context.Context,
	db database.IDB,
	appID string,
	extraOpts ...bunex.SelectQueryOption,
) ([]*entity.App, error) {
	listOpts := []bunex.SelectQueryOption{
		bunex.SelectWhere("app.parent_id = ?", appID),
	}
	listOpts = append(listOpts, extraOpts...)

	apps, _, err := s.appRepo.List(ctx, db, "", nil, listOpts...)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	return apps, nil
}
