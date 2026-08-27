package envvarserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
)

func (s *service) BuildSharedEnvVarsInApp(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	buildOptions envvarservice.EnvBuildOptions,
) ([]*envvarservice.EnvVar, error) {
	buildOptions.SharedVarsOnly = true
	buildOptions.SkipLoadingSecrets = true
	resp, err := s.BuildEnvVarsInApp(ctx, db, &envvarservice.BuildEnvVarsInAppReq{
		App:          app,
		BuildOptions: buildOptions,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return resp.EnvVars, nil
}

func (s *service) buildSharedEnvVarsInApp(
	ctx context.Context,
	db database.IDB,
	projectID string,
	projectEnvID string,
	appKey string,
	buildOptions envvarservice.EnvBuildOptions,
) ([]*envvarservice.EnvVar, error) {
	listOpts := []bunex.SelectQueryOption{
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		// bunex.SelectWhere("app.status = ?", base.AppStatusActive),
		bunex.SelectWhere("app.key = ?", appKey),
		bunex.SelectWhere("app.project_env_id = ?", projectEnvID),
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
		bunex.SelectRelation("ProjectEnv"),
	}
	apps, _, err := s.appRepo.List(ctx, db, projectID, nil, listOpts...)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if len(apps) == 0 {
		return nil, hperrors.NewNotFound("App")
	}
	return s.BuildSharedEnvVarsInApp(ctx, db, apps[0], buildOptions)
}
