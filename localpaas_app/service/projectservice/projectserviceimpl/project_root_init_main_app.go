package projectserviceimpl

import (
	"context"
	"strings"

	"github.com/moby/moby/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/config"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/contenttypes"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/ulid"
	"github.com/localpaas/localpaas/localpaas_app/pkg/unit"
)

var (
	localpaasAppInitExcludedEnvs = map[string]struct{}{
		"LP_USER_ADMIN_USERNAME": {},
		"LP_USER_ADMIN_PASSWORD": {},
		"LP_USER_ADMIN_EMAIL":    {},
	}
)

func (s *service) initRootProjectMainApp(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	service *swarm.Service,
) (shouldUpdateService bool, err error) {
	timeNow := timeutil.NowUTC()
	cfg := config.Current

	// Add service settings for the app
	dbServiceSetting := &entity.Setting{
		ID:        gofn.Must(ulid.NewStringULID()),
		Scope:     base.SettingScopeGlobal,
		Type:      base.SettingTypeLocalPaaSService,
		Status:    base.SettingStatusActive,
		Name:      "Service settings",
		Version:   entity.CurrentLocalPaaSServiceVersion,
		CreatedAt: timeNow,
		UpdatedAt: timeNow,
	}
	serviceSetting := &entity.LocalPaaSService{
		AppSettings: entity.LocalPaaSAppSettings{
			Replicas: 1,
		},
		WorkerSettings: entity.LocalPaaSWorkerSettings{
			Replicas:           0,
			Concurrency:        cfg.Tasks.Queue.Concurrency,
			RunWorkerInMainApp: true,
		},
		TaskSettings: entity.LocalPaaSTaskSettings{
			TaskCheckInterval:  timeutil.Duration(cfg.Tasks.Queue.TaskCheckInterval),
			TaskCreateInterval: timeutil.Duration(cfg.Tasks.Queue.TaskCreateInterval),
		},
		HealthcheckSettings: entity.LocalPaaSHealthcheckSettings{
			BaseInterval: timeutil.Duration(cfg.Tasks.Healthcheck.BaseInterval),
		},
	}
	dbServiceSetting.MustSetData(serviceSetting)

	// Add HTTP settings for the main app
	dbHttpSetting := &entity.Setting{
		ID:        gofn.Must(ulid.NewStringULID()),
		Scope:     base.SettingScopeApp,
		ObjectID:  app.ID,
		Type:      base.SettingTypeAppHttp,
		Status:    base.SettingStatusActive,
		Version:   entity.CurrentAppHttpSettingsVersion,
		CreatedAt: timeNow,
		UpdatedAt: timeNow,
	}
	httpSettings := &entity.AppHttpSettings{
		ExposePublicly: true,
		Domains: []*entity.AppDomain{
			{
				Enabled:       true,
				Domain:        cfg.AppDomain,
				ContainerPort: cfg.HTTPServer.Port,
				ForceHttps:    true,
				CompressionConfig: &entity.HTTPCompressionConfig{
					Enabled:              true,
					IncludedContentTypes: contenttypes.ContentTypesShouldCompress,
					MinResponseBody:      unit.KB, // 1kb
					DefaultEncoding:      "br",    // brotli
				},
			},
		},
	}
	dbHttpSetting.MustSetData(httpSettings)

	// Sync env-vars from the swarm service
	dbEnvVarsSetting := &entity.Setting{
		ID:        gofn.Must(ulid.NewStringULID()),
		Scope:     base.SettingScopeApp,
		ObjectID:  app.ID,
		Type:      base.SettingTypeEnvVar,
		Status:    base.SettingStatusActive,
		Version:   entity.CurrentEnvVarsVersion,
		CreatedAt: timeNow,
		UpdatedAt: timeNow,
	}
	envVars := &entity.EnvVars{}
	var newEnv []string
	for _, env := range service.Spec.TaskTemplate.ContainerSpec.Env {
		k, v, _ := strings.Cut(env, "=")
		if _, exists := localpaasAppInitExcludedEnvs[k]; exists {
			shouldUpdateService = true
			continue
		}
		newEnv = append(newEnv, env)
		envVars.Data = append(envVars.Data, &entity.EnvVar{
			Key:       k,
			Value:     v,
			IsLiteral: true,
		})
	}
	if shouldUpdateService {
		service.Spec.TaskTemplate.ContainerSpec.Env = newEnv
	}

	dbEnvVarsSetting.MustSetData(envVars)

	// Insert the settings into DB
	err = s.settingRepo.InsertMulti(ctx, db, []*entity.Setting{dbServiceSetting, dbHttpSetting, dbEnvVarsSetting})
	if err != nil {
		return false, apperrors.Wrap(err)
	}

	return shouldUpdateService, nil
}
