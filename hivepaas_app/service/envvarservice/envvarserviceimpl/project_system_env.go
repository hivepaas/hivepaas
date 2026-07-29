package envvarserviceimpl

import (
	"context"
	"sort"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
)

func (s *service) BuildSystemEnvVarsInProject(
	ctx context.Context,
	_ database.IDB,
	req *envvarservice.BuildSystemEnvVarsInProjectReq,
) ([]*envvarservice.EnvVar, error) {
	result := []*envvarservice.EnvVar{
		{
			EnvVar: &entity.EnvVar{
				Key:      base.ProjectSystemEnvVarName,
				Value:    req.Project.Name,
				IsShared: true,
			},
		},
		{
			EnvVar: &entity.EnvVar{
				Key:      base.ProjectSystemEnvVarID,
				Value:    req.Project.ID,
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
