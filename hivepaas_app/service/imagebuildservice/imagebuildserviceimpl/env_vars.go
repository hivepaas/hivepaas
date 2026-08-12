package imagebuildserviceimpl

import (
	"context"
	"os"
	"strings"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
)

func (s *service) calcBuildEnvVars(
	ctx context.Context,
	db database.IDB,
	data *imageBuildData,
) (map[string]*string, error) {
	envResp, err := s.envVarService.BuildEnvVarsInApp(ctx, db, &envvarservice.BuildEnvVarsInAppReq{
		App: data.App,
		LoadOptions: envvarservice.EnvLoadOptions{
			BuildPhase: true,
		},
		BuildOptions: envvarservice.EnvBuildOptions{
			BuildPhaseOnly: true,
		},
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	if len(envResp.EnvVars) > 0 {
		secrets := make(map[string]struct{}, 10) //nolint:mnd
		for _, env := range envResp.EnvVars {
			for secret := range env.RefSecrets {
				plainSecret, err := secret.Value.GetPlain()
				if err != nil {
					return nil, apperrors.Wrap(err)
				}
				secrets[plainSecret] = struct{}{}
			}
		}
		data.LogStore.UpdateRedactorAddSecrets(gofn.MapKeys(secrets))
	}

	result := make(map[string]*string, len(envResp.EnvVars))
	for _, envVar := range envResp.EnvVars {
		result[envVar.Key] = &envVar.Value
	}

	return result, nil
}

func (s *service) calcSafeEnvVars() []string {
	envs := make([]string, 0, 20) //nolint:mnd
	for _, env := range os.Environ() {
		if !strings.HasPrefix(env, "HP_") {
			envs = append(envs, env)
		}
	}
	return envs
}
