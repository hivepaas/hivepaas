package envvarserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
)

func (s *service) ComputeAppSharedEnvVars(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	buildPhase bool,
	skipLoadingSecrets bool,
	maskSecrets bool,
) ([]*envvarservice.EnvVar, error) {
	return s.ComputeAppEnvVars(ctx, db, &envvarservice.ComputeAppEnvVarsReq{
		App:                app,
		SkipLoadingSecrets: skipLoadingSecrets,
		MaskSecrets:        maskSecrets,
		BuildPhaseOnly:     buildPhase,
		SharedVarsOnly:     true,
	})
}

func (s *service) computeAppSharedEnvVars(
	ctx context.Context,
	db database.IDB,
	projectID string,
	projectEnvID string,
	appKey string,
	buildPhase bool,
	skipLoadingSecrets bool,
	maskSecrets bool,
) ([]*envvarservice.EnvVar, error) {
	listOpts := []bunex.SelectQueryOption{
		// bunex.SelectWhere("app.status = ?", base.AppStatusActive),
		bunex.SelectWhere("app.key = ?", appKey),
		bunex.SelectWhere("app.project_env_id = ?", projectEnvID),
	}
	apps, _, err := s.appRepo.List(ctx, db, projectID, nil, listOpts...)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	if len(apps) == 0 {
		return nil, apperrors.NewNotFound("App")
	}
	return s.ComputeAppSharedEnvVars(ctx, db, apps[0], buildPhase, skipLoadingSecrets, maskSecrets)
}
