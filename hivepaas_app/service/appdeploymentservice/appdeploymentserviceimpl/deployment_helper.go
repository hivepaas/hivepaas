package appdeploymentserviceimpl

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
)

func (s *service) loadImageBuildSettings(
	ctx context.Context,
	db database.IDB,
	data *repoDeploymentData,
) error {
	setting, err := s.settingRepo.GetSingle(ctx, db, data.App.Project.GetObjectScope(),
		base.SettingTypeImageBuildSettings, true)
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return apperrors.Wrap(err)
	}
	if setting != nil {
		data.ImageBuildSettings = setting.MustAsImageBuildSettings()
	}
	return nil
}

func (s *service) reloadApp(
	ctx context.Context,
	db database.IDB,
	requireProjectActive bool,
	requireAppActive bool,
	data *appDeploymentData,
) (err error) {
	if data == nil || data.App == nil {
		return nil
	}

	app, err := s.appService.LoadApp(ctx, db, data.App.ProjectID, data.App.ID,
		requireProjectActive, requireAppActive,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
		bunex.SelectRelation("ProjectEnv"),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}
	data.App = app

	return nil
}
