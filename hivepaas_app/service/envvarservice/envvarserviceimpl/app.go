package envvarserviceimpl

import (
	"context"
	"maps"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
)

const (
	refSecretMaxSize = 10 * 1024 // 10 KB
)

func (s *service) BuildAppEnvVars(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	buildPhase bool,
) (res []*envvarservice.EnvVar, usedSecrets []*entity.Secret, err error) {
	vars, secrets, err := s.LoadAppEnvVarsAndSecrets(ctx, db, app, true, true, buildPhase)
	if err != nil {
		return nil, nil, apperrors.New(err)
	}

	// App inherits ENV vars in order from the parent app, then the project, and then from global
	envStore := vars.FinalEnvVars()
	secretStore := secrets.FinalSecrets()

	// Construct result
	res = make([]*envvarservice.EnvVar, 0, len(envStore))
	for _, env := range envStore {
		res = append(res, &envvarservice.EnvVar{EnvVar: env})
	}

	// Process all references within the ENV values
	usedSecrets = make([]*entity.Secret, 0)
	for _, env := range res {
		if env.IsLiteral {
			continue
		}
		s.processRefs(env, envStore, secretStore, &usedSecrets)
	}

	return res, usedSecrets, nil
}

func (s *service) ProcessEnvRefs(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	envVars []*entity.EnvVar,
	loadEnvVars bool,
	loadSecrets bool,
	buildPhase bool,
) (res []*envvarservice.EnvVar, usedSecrets []*entity.Secret, err error) {
	if len(envVars) == 0 {
		return nil, nil, nil
	}
	vars, secrets, err := s.LoadAppEnvVarsAndSecrets(ctx, db, app, loadEnvVars, loadSecrets, buildPhase)
	if err != nil {
		return nil, nil, apperrors.New(err)
	}

	// Construct result
	res = make([]*envvarservice.EnvVar, 0, len(envVars))
	for _, env := range envVars {
		res = append(res, &envvarservice.EnvVar{EnvVar: env})
	}

	// App inherits ENV vars in order from the parent app, then the project, and then from global
	envStore := vars.FinalEnvVars()
	if envStore == nil {
		envStore = map[string]*entity.EnvVar{}
	}
	// Update the envStore with the input values
	for _, env := range envVars {
		envStore[env.Key] = env
	}
	secretStore := secrets.FinalSecrets()
	if secretStore == nil {
		secretStore = map[string]*entity.Secret{}
	}
	// Process all references within the ENV values
	usedSecrets = make([]*entity.Secret, 0)
	for _, env := range res {
		if env.IsLiteral {
			continue
		}
		s.processRefs(env, envStore, secretStore, &usedSecrets)
	}
	return res, usedSecrets, nil
}

type AppEnvVars struct {
	App       map[string]*entity.EnvVar
	ParentApp map[string]*entity.EnvVar
	Project   map[string]*entity.EnvVar
	Global    map[string]*entity.EnvVar
}

func (ev *AppEnvVars) FinalEnvVars() map[string]*entity.EnvVar {
	if ev == nil {
		return nil
	}
	res := make(map[string]*entity.EnvVar, len(ev.App)+10) //nolint
	maps.Copy(res, ev.Global)
	maps.Copy(res, ev.Project)
	maps.Copy(res, ev.ParentApp)
	maps.Copy(res, ev.App)
	return res
}

type AppSecrets struct {
	App       map[string]*entity.Secret
	ParentApp map[string]*entity.Secret
	Project   map[string]*entity.Secret
	Global    map[string]*entity.Secret
}

func (ev *AppSecrets) FinalSecrets() map[string]*entity.Secret {
	if ev == nil {
		return nil
	}
	res := make(map[string]*entity.Secret, len(ev.App)+10) //nolint
	maps.Copy(res, ev.Global)
	maps.Copy(res, ev.Project)
	maps.Copy(res, ev.ParentApp)
	maps.Copy(res, ev.App)
	return res
}

//nolint:gocognit
func (s *service) LoadAppEnvVarsAndSecrets(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	loadEnvVars bool,
	loadSecrets bool,
	buildPhase bool,
) (envVars *AppEnvVars, secrets *AppSecrets, err error) {
	if !loadEnvVars && !loadSecrets {
		return nil, nil, nil
	}
	if loadEnvVars {
		//nolint:mnd
		envVars = &AppEnvVars{
			App:       make(map[string]*entity.EnvVar, 20),
			ParentApp: make(map[string]*entity.EnvVar, 20),
			Project:   make(map[string]*entity.EnvVar, 20),
			Global:    make(map[string]*entity.EnvVar, 20),
		}
	}
	if loadSecrets {
		//nolint:mnd
		secrets = &AppSecrets{
			App:       make(map[string]*entity.Secret, 10),
			ParentApp: make(map[string]*entity.Secret, 10),
			Project:   make(map[string]*entity.Secret, 10),
			Global:    make(map[string]*entity.Secret, 10),
		}
	}

	settings, _, err := s.settingRepo.List(ctx, db, app.GetObjectScope(), nil,
		bunex.SelectWhereGroup(
			bunex.SelectWhereIf(loadEnvVars, "setting.type = ?", base.SettingTypeEnvVar),
			bunex.SelectWhereIf(!loadEnvVars, "1=1"),
			bunex.SelectWhereOrIf(loadSecrets, "(setting.type = ? AND setting.size <= ?)",
				base.SettingTypeSecret, refSecretMaxSize),
		),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
	)
	if err != nil {
		return nil, nil, apperrors.New(err)
	}

	if len(settings) == 0 {
		return envVars, secrets, nil
	}

	for _, setting := range settings {
		if setting.Type == base.SettingTypeEnvVar {
			vars := setting.MustAsEnvVars()
			switch {
			case setting.ObjectID == app.ID:
				for _, env := range vars.Data {
					if env.IsBuildEnv == buildPhase {
						envVars.App[env.Key] = env
					}
				}
			case app.ParentID != "" && setting.ObjectID == app.ParentID:
				for _, env := range vars.Data {
					if env.IsBuildEnv == buildPhase {
						envVars.ParentApp[env.Key] = env
					}
				}
			case setting.ObjectID == "":
				for _, env := range vars.Data {
					if env.IsBuildEnv == buildPhase {
						envVars.Global[env.Key] = env
					}
				}
			default:
				for _, env := range vars.Data {
					if env.IsBuildEnv == buildPhase {
						envVars.Project[env.Key] = env
					}
				}
			}
			continue
		}

		if setting.Type == base.SettingTypeSecret {
			secret := setting.MustAsSecret() // decryption takes time, so do it when needed only
			switch {
			case setting.ObjectID == app.ID:
				secrets.App[setting.Name] = secret
			case app.ParentID != "" && setting.ObjectID == app.ParentID:
				secrets.ParentApp[setting.Name] = secret
			case setting.ObjectID == "":
				secrets.Global[setting.Name] = secret
			default:
				secrets.Project[setting.Name] = secret
			}
		}
	}

	return envVars, secrets, nil
}
