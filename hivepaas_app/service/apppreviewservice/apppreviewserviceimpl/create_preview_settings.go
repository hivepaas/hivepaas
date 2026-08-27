package apppreviewserviceimpl

import (
	"context"
	"fmt"
	"strings"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

func (s *service) onCloneApp(
	targetApp, srcApp *entity.App,
	data *createPreviewData,
) error {
	targetApp.Name = data.CalcAppName
	targetApp.ProjectEnvID = data.App.ProjectEnvID
	targetApp.ProjectEnv = data.App.ProjectEnv
	targetApp.Status = base.AppStatusActive
	targetApp.ParentID = srcApp.ID // Preview app must be a child app of the current
	targetApp.ParentApp = srcApp
	return nil
}

func (s *service) onCloneAppSetting(
	ctx context.Context,
	db database.IDB,
	setting *entity.Setting,
	data *createPreviewData,
) (*entity.Setting, error) {
	switch setting.Type { //nolint:exhaustive
	case base.SettingTypeApp:
		return nil, nil
	case base.SettingTypeAppDeployment:
		return s.onCloneDeploymentSetting(setting, data)
	case base.SettingTypeAppRouting:
		return s.onCloneRoutingSetting(ctx, db, setting, data)
	case base.SettingTypeEnvVar:
		return s.onCloneEnvVars(setting, data)
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

func (s *service) onCloneDeploymentSetting(
	setting *entity.Setting,
	data *createPreviewData,
) (*entity.Setting, error) {
	deploymentSettings := setting.MustAsAppDeploymentSettings()
	deploymentSettings.RepoSource.RepoRef = data.CalcRepoRef
	deploymentSettings.RepoSource.CommitHash = "" // unset target commit
	data.DeploymentSettings = deploymentSettings

	setting.MustSetData(deploymentSettings)
	return setting, nil
}

func (s *service) onCloneRoutingSetting(
	ctx context.Context,
	db database.IDB,
	setting *entity.Setting,
	data *createPreviewData,
) (*entity.Setting, error) {
	routingSettings := setting.MustAsAppRoutingSettings()

	var activeDomains []string
	currDomains := routingSettings.Domains
	routingSettings.Domains = nil
	for _, domain := range currDomains {
		if !domain.Enabled {
			continue
		}
		subdomain := strings.TrimSuffix(data.CalcSubdomain, "."+domain.Domain)
		domain.Domain = fmt.Sprintf("%v.%v", subdomain, domain.Domain)
		// TODO: handle SSL cert
		routingSettings.Domains = append(routingSettings.Domains, domain)
		activeDomains = append(activeDomains, domain.Domain)
	}

	// Make sure all domains used by the app are not hold by any other app
	err := s.domainService.VerifyDomainsAvailable(ctx, db, activeDomains, []string{data.PreviewApp.ID})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	setting.MustSetData(routingSettings)
	return setting, nil
}

func (s *service) onCloneEnvVars(
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

func (s *service) onCloneAppService(
	targetSvc, _ *swarm.Service,
	data *createPreviewData,
) error {
	targetSvcSpec := &targetSvc.Spec
	if targetSvcSpec.Mode.Replicated != nil {
		targetSvcSpec.Mode.Replicated.Replicas = new(uint64(1))
	}

	// When we need to clone related DB apps, must not start the preview
	noStartService := data.Args.NoStart || len(data.CloneDBApps) > 0

	if noStartService { // If noStart, use replicated service mode with replicas = 0
		if targetSvcSpec.Mode.Replicated == nil {
			targetSvcSpec.Mode = swarm.ServiceMode{
				Replicated: &swarm.ReplicatedService{},
			}
		}
		targetSvcSpec.Mode.Replicated.Replicas = new(uint64(0))
	}
	return nil
}
