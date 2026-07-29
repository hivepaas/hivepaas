package envvarserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
)

func (s *service) DefaultEnvLoad(
	ctx context.Context,
	db database.IDB,
	scope *base.ObjectScope,
	options envvarservice.EnvLoadOptions,
) (envVars []*envvarservice.EnvVar, secrets []*entity.Setting, err error) {
	settings, _, err := s.settingRepo.List(ctx, db, nil, nil,
		bunex.SelectWhereGroup(
			bunex.SelectWhere("setting.type = ?", base.SettingTypeEnvVar),
			bunex.SelectWhereOr("(setting.type = ? AND setting.size <= ?)",
				base.SettingTypeSecret, refSecretMaxSize),
		),
		bunex.SelectWhere("setting.object_id = ?", scope.ScopeObjectID()),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
	)
	if err != nil {
		return nil, nil, apperrors.Wrap(err)
	}

	envVars = make([]*envvarservice.EnvVar, 0, 20) //nolint:mnd
	secrets = make([]*entity.Setting, 0, 10)       //nolint:mnd
	for _, setting := range settings {
		switch setting.Type { //nolint:exhaustive
		case base.SettingTypeEnvVar:
			for _, env := range setting.MustAsEnvVars().Data {
				if env.IsBuild == options.BuildPhase {
					envVars = append(envVars, &envvarservice.EnvVar{EnvVar: env})
				}
			}
		case base.SettingTypeSecret:
			secrets = append(secrets, setting)
		}
	}

	return envVars, secrets, nil
}
