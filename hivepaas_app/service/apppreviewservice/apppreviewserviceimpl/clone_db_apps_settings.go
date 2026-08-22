package apppreviewserviceimpl

import (
	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

func (s *service) onCloneDBApp(
	targetApp, srcApp *entity.App,
	data *createPreviewData,
) error {
	appData := data.CloneDBAppsData[srcApp.ID]
	targetApp.Name = appData.NewAppName
	targetApp.Key = appData.NewAppKey
	targetApp.ProjectEnvID = srcApp.ProjectEnvID
	targetApp.ProjectEnv = srcApp.ProjectEnv
	targetApp.Status = base.AppStatusActive
	targetApp.ParentID = srcApp.ID // Preview app must be a child app of the current
	targetApp.ParentApp = srcApp
	return nil
}

func (s *service) onCloneDBAppSetting(
	setting *entity.Setting,
	data *createPreviewData,
) (*entity.Setting, error) {
	switch setting.Type { //nolint:exhaustive
	case base.SettingTypeApp:
		return nil, nil
	case base.SettingTypeAppDeployment:
		return nil, nil
	case base.SettingTypeAppRouting:
		return nil, nil
	case base.SettingTypeEnvVar:
		return s.onCloneDBAppEnvVars(setting, data)
	case base.SettingTypeSecret:
		return nil, nil
	case base.SettingTypeConfigFile:
		return nil, nil
	case base.SettingTypePeriodicJob:
		return nil, nil
	case base.SettingTypeSchedJob:
		return nil, nil
	case base.SettingTypeAppFeatures:
		return nil, nil
	default:
		return nil, nil
	}
}

func (s *service) onCloneDBAppEnvVars(
	setting *entity.Setting,
	data *createPreviewData,
) (*entity.Setting, error) {
	envVars := setting.MustAsEnvVars()
	changedEnvVars := make([]*entity.EnvVar, 0)
	for _, env := range envVars.Data {
		if env.IsLiteral {
			continue
		}
		// E.g. In `backend` app there is env `CONNECT_STR=scheme://${db.HOST}:${db.PORT}`,
		// it will be converted to `CONNECT_STR=scheme://${cloned_db.HOST}:${cloned_db.PORT}`.
		newValue := data.CloneDBAppsEnvRefReplacer.Replace(env.Value)
		if newValue != env.Value {
			env.Value = newValue
			changedEnvVars = append(changedEnvVars, env)
		}
	}
	if len(changedEnvVars) == 0 {
		return nil, nil
	}
	envVars.Data = changedEnvVars
	setting.MustSetData(envVars)
	return setting, nil
}

func (s *service) onCloneDBAppService(
	targetSvc, _ *swarm.Service,
) error {
	// Set the replicas of the cloned app to 1
	if targetSvc.Spec.Mode.Replicated != nil {
		targetSvc.Spec.Mode.Replicated.Replicas = new(uint64(1))
	}
	return nil
}
