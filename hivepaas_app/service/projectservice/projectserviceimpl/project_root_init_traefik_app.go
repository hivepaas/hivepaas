package projectserviceimpl

import (
	"context"
	"strings"

	"github.com/moby/moby/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
)

var (
	traefikAppInitExcludedEnvs = map[string]struct{}{}
)

func (s *service) initRootProjectTraefikApp(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	service *swarm.Service,
) (shouldUpdateService bool, err error) {
	timeNow := timeutil.NowUTC()

	// Add service settings for the app
	dbServiceSetting := &entity.Setting{
		ID:        gofn.Must(ulid.NewStringULID()),
		Scope:     base.ObjectScopeGlobal,
		Type:      base.SettingTypeTraefikService,
		Status:    base.SettingStatusActive,
		Name:      "Service settings",
		Version:   entity.CurrentTraefikServiceVersion,
		CreatedAt: timeNow,
		UpdatedAt: timeNow,
	}
	serviceSetting := &entity.TraefikService{
		AppSettings: entity.TraefikAppSettings{
			Replicas: 1,
		},
	}
	dbServiceSetting.MustSetData(serviceSetting)

	// Sync env-vars from the swarm service
	dbEnvVarsSetting := &entity.Setting{
		ID:        gofn.Must(ulid.NewStringULID()),
		Scope:     base.ObjectScopeApp,
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
		if _, exists := traefikAppInitExcludedEnvs[k]; exists {
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
	err = s.settingRepo.InsertMulti(ctx, db, []*entity.Setting{dbServiceSetting, dbEnvVarsSetting})
	if err != nil {
		return false, apperrors.Wrap(err)
	}

	return shouldUpdateService, nil
}
