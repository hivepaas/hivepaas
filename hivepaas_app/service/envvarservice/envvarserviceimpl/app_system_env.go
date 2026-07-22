package envvarserviceimpl

import (
	"context"
	"sort"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
)

func (s *service) ComputeAppSystemEnvVars(
	ctx context.Context,
	req *envvarservice.ComputeAppSystemEnvVarsReq,
) ([]*envvarservice.EnvVar, error) {
	result := []*envvarservice.EnvVar{
		{
			EnvVar: &entity.EnvVar{
				Key:      base.AppSystemEnvVarHost,
				Value:    req.App.Key,
				IsShared: true,
			},
		},
		{
			EnvVar: &entity.EnvVar{
				Key:      base.AppSystemEnvVarName,
				Value:    req.App.Name,
				IsShared: true,
			},
		},
		{
			EnvVar: &entity.EnvVar{
				Key:      base.AppSystemEnvVarID,
				Value:    req.App.ID,
				IsShared: true,
			},
		},
		{
			EnvVar: &entity.EnvVar{
				Key:      base.AppSystemEnvVarEnv,
				Value:    req.App.Env,
				IsShared: true,
			},
		},
	}

	for _, env := range result {
		env.IsLiteral = true
		env.IsSystem = true
	}

	if req.Sort {
		sort.Slice(result, func(i, j int) bool {
			return result[i].Key < result[j].Key
		})
	}

	return result, nil
}
