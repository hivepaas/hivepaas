package imagebuildserviceimpl

import (
	"context"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/envutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
)

// calcBuildEnvVars is an app's build variables, and the secret values in them.
func (s *service) calcBuildEnvVars(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
) (envVars map[string]*string, secretValues []string, err error) {
	envResp, err := s.envVarService.BuildEnvVarsInApp(ctx, db, &envvarservice.BuildEnvVarsInAppReq{
		App: app,
		LoadOptions: envvarservice.EnvLoadOptions{
			BuildPhase: true,
		},
		BuildOptions: envvarservice.EnvBuildOptions{
			BuildPhaseOnly: true,
		},
	})
	if err != nil {
		return nil, nil, hperrors.Wrap(err)
	}

	secrets := make(map[string]struct{}, 10) //nolint:mnd
	for _, env := range envResp.EnvVars {
		for secret := range env.RefSecrets {
			plainSecret, err := secret.Value.GetPlain()
			if err != nil {
				return nil, nil, hperrors.Wrap(err)
			}
			secrets[plainSecret] = struct{}{}
		}
	}

	envVars = make(map[string]*string, len(envResp.EnvVars))
	for _, envVar := range envResp.EnvVars {
		envVars[envVar.Key] = &envVar.Value
	}

	return envVars, gofn.MapKeys(secrets), nil
}

func (s *service) calcSafeEnvVars() []string {
	return envutil.SafeEnviron()
}
