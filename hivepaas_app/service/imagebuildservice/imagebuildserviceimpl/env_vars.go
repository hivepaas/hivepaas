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

// calcBuildEnvVars is an app's build variables - those that use no secret, and
// those that do - and the secret values in them.
func (s *service) calcBuildEnvVars(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
) (envVars map[string]*string, secretEnvVars map[string]string, secretValues []string, err error) {
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
		return nil, nil, nil, hperrors.Wrap(err)
	}

	secrets := make(map[string]struct{}, 10) //nolint:mnd
	envVars = make(map[string]*string, len(envResp.EnvVars))
	secretEnvVars = make(map[string]string)
	for _, envVar := range envResp.EnvVars {
		for secret := range envVar.RefSecrets {
			plainSecret, err := secret.Value.GetPlain()
			if err != nil {
				return nil, nil, nil, hperrors.Wrap(err)
			}
			secrets[plainSecret] = struct{}{}
		}
		// A value with a secret anywhere in it is a secret as a whole.
		if len(envVar.RefSecrets) > 0 {
			secretEnvVars[envVar.Key] = envVar.Value
			continue
		}
		envVars[envVar.Key] = &envVar.Value
	}

	return envVars, secretEnvVars, gofn.MapKeys(secrets), nil
}

func (s *service) calcSafeEnvVars() []string {
	return envutil.SafeEnviron()
}
