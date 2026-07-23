package envvarserviceimpl

import (
	"context"
	"sort"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
)

func (s *service) ComputeAppSystemEnvVars(
	ctx context.Context,
	db database.IDB,
	req *envvarservice.ComputeAppSystemEnvVarsReq,
) ([]*envvarservice.EnvVar, error) {
	httpLinks, _, err := s.resLinkRepo.List(ctx, db, nil,
		bunex.SelectJoin("JOIN settings ON settings.id = res_link.src_id"),
		bunex.SelectWhere("res_link.src_type = ?", base.ResourceTypeSetting),
		bunex.SelectWhere("settings.object_id = ?", req.App.ID),
		bunex.SelectWhereIn("res_link.dst_type IN (?)", base.ResourceTypePort),
	)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	httpResLinks := entity.ResLinks(httpLinks)

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
				Key:      base.AppSystemEnvVarPort,
				Value:    gofn.Head(httpResLinks.GetDstIDByDstType(base.ResourceTypePort)),
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
