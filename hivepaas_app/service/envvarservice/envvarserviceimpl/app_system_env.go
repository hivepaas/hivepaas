package envvarserviceimpl

import (
	"context"
	"sort"
	"strconv"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/settinghelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
)

//nolint:funlen
func (s *service) BuildSystemEnvVarsInApp(
	ctx context.Context,
	db database.IDB,
	req *envvarservice.BuildSystemEnvVarsInAppReq,
) ([]*envvarservice.EnvVar, error) {
	settings, _, err := s.settingRepo.List(ctx, db, nil, nil,
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
		bunex.SelectWhereIn("setting.type IN (?)", base.SettingTypeAppRouting, base.SettingTypeAppKind),
		bunex.SelectWhere("setting.object_id = ?", req.App.ID),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	routingSetting := settinghelper.FindSettingByType(settings, base.SettingTypeAppRouting)
	var routingSettings *entity.AppRoutingSettings
	if routingSetting != nil {
		routingSettings = routingSetting.MustAsAppRoutingSettings()
	}

	portStr := ""
	activeDomain := ""
	if routingSettings != nil {
		if routingSettings.Port > 0 {
			portStr = strconv.Itoa(routingSettings.Port)
		}
		activeDomain = gofn.FirstOr(routingSettings.GetActiveDomainNames(), "")
	}

	kindSetting := settinghelper.FindSettingByType(settings, base.SettingTypeAppKind)
	var kindSettings *entity.AppKindSettings
	if kindSetting != nil {
		kindSettings = kindSetting.MustAsAppKindSettings()
	}

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
				Value:    portStr,
				IsShared: true,
			},
		},
		{
			EnvVar: &entity.EnvVar{
				Key:      base.AppSystemEnvVarDomain,
				Value:    activeDomain,
				IsShared: true,
			},
		},
		{
			EnvVar: &entity.EnvVar{
				Key:      base.AppSystemEnvVarEnv,
				Value:    req.App.ProjectEnv.Name,
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
	}

	if kindSettings != nil && kindSettings.Category == base.AppCategoryDatabase && kindSettings.Database != nil {
		if err := kindSettings.Decrypt(); err != nil {
			return nil, hperrors.Wrap(err)
		}
		result = append(result,
			&envvarservice.EnvVar{
				EnvVar: &entity.EnvVar{
					Key:      base.AppSystemEnvVarUser,
					Value:    kindSettings.Database.Username,
					IsShared: true,
				},
			},
			&envvarservice.EnvVar{
				EnvVar: &entity.EnvVar{
					Key:      base.AppSystemEnvVarPassword,
					Value:    gofn.Must(kindSettings.Database.Password.GetPlain()),
					IsShared: true,
				},
			},
			&envvarservice.EnvVar{
				EnvVar: &entity.EnvVar{
					Key:      base.AppSystemEnvVarRootPassword,
					Value:    gofn.Must(kindSettings.Database.RootPassword.GetPlain()),
					IsShared: false, // NOTE: do not share this env with other apps
				},
			},
			&envvarservice.EnvVar{
				EnvVar: &entity.EnvVar{
					Key:      base.AppSystemEnvVarDatabaseName,
					Value:    kindSettings.Database.DbName,
					IsShared: true,
				},
			},
			&envvarservice.EnvVar{
				EnvVar: &entity.EnvVar{
					Key:      base.AppSystemEnvVarSSLMode,
					Value:    string(kindSettings.Database.SSLMode),
					IsShared: true,
				},
			})
	}

	// TODO: envs for cache app

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
