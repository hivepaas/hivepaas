package appcloneserviceimpl

import (
	"context"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
)

const (
	dockerImageInit    = "busybox:latest"
	dockerImageInitDev = "crccheck/hello-world:latest"
)

func (s *service) cloneAppSettings(
	ctx context.Context,
	db database.IDB,
	data *appCloneData,
) (err error) {
	destApp, srcApp := data.DestApp, data.SrcApp
	appSettings, _, err := s.settingRepo.List(ctx, db, nil, nil,
		bunex.SelectWhere("setting.scope = ?", base.ObjectScopeApp),
		bunex.SelectWhere("setting.object_id = ?", srcApp.ID),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}

	cloneFunc := data.OnCloneSetting
	if cloneFunc == nil {
		cloneFunc = func(destApp, srcApp *entity.App, setting *entity.Setting) (*entity.Setting, error) {
			return s.onCloneSettingDefault(setting, data)
		}
	}

	for _, setting := range appSettings {
		cpSetting, err := setting.Clone(true)
		if err != nil {
			return apperrors.Wrap(err)
		}
		cpSetting.ObjectID = destApp.ID
		cpSetting.CreatedAt = data.TimeNow
		cpSetting.UpdatedAt = data.TimeNow
		cpSetting.UpdateVer = 0
		st, err := cloneFunc(destApp, srcApp, cpSetting)
		if err != nil {
			return apperrors.Wrap(err)
		}
		if st != nil {
			data.ClonedSettings = append(data.ClonedSettings, st)
		}
	}

	destApp.Settings = data.ClonedSettings

	// Update ref app for every sched job
	for _, jobSetting := range destApp.GetSettingsByType(base.SettingTypeSchedJob) {
		schedJob := jobSetting.MustAsSchedJob()
		schedJob.App.ID = destApp.ID
		jobSetting.MustSetData(schedJob)
	}

	// Validation

	// Active domains of the app need to validate
	newHttpSetting := destApp.GetSettingByType(base.SettingTypeAppRouting)
	if newHttpSetting != nil {
		activeDomains := newHttpSetting.MustAsAppRoutingSettings().GetActiveDomainNames()

		// Verify domains are allowed in project
		err = s.domainService.VerifyProjectDomains(ctx, db, destApp.ProjectID, activeDomains)
		if err != nil {
			return apperrors.Wrap(err)
		}

		// Make sure all domains used by the app are not hold by any other app
		err = s.domainService.VerifyDomainsAvailable(ctx, db, activeDomains, []string{destApp.ID})
		if err != nil {
			return apperrors.Wrap(err)
		}
	}

	return nil
}

func (s *service) onCloneSettingDefault(
	setting *entity.Setting,
	data *appCloneData,
) (*entity.Setting, error) {
	settings := data.CloneSettings
	switch setting.Type { //nolint:exhaustive
	case base.SettingTypeApp:
		return setting, nil
	case base.SettingTypeAppDeployment:
		return s.onCloneDeploymentSettingDefault(setting, data)
	case base.SettingTypeAppRouting:
		return s.onCloneHttpSettingDefault(setting, data)
	case base.SettingTypeAppFeatures:
		return setting, nil
	case base.SettingTypeEnvVar:
		return gofn.If(settings.CloneEnvVars, setting, nil), nil
	case base.SettingTypeSecret:
		return gofn.If(settings.CloneSecrets, setting, nil), nil
	case base.SettingTypeConfigFile:
		return gofn.If(settings.CloneConfigFiles, setting, nil), nil
	case base.SettingTypePeriodicJob:
		return gofn.If(settings.ClonePeriodicJobs, setting, nil), nil
	case base.SettingTypeSchedJob:
		return gofn.If(settings.CloneSchedJobs, setting, nil), nil
	default:
		return nil, nil
	}
}

func (s *service) onCloneDeploymentSettingDefault(
	setting *entity.Setting,
	data *appCloneData,
) (*entity.Setting, error) {
	settings := data.CloneSettings
	deploymentSettings := setting.MustAsAppDeploymentSettings()

	if !settings.CloneDeploymentSettings {
		isDevEnv := config.Current.IsDevEnv()
		deploymentSettings.ActiveMethod = base.DeploymentMethodImage
		deploymentSettings.ImageSource = &entity.DeploymentImageSource{
			Image: gofn.If(isDevEnv, dockerImageInitDev, dockerImageInit),
		}
		deploymentSettings.Command = gofn.If(isDevEnv, "sleep infinity", "")
		deploymentSettings.WorkingDir = ""
		deploymentSettings.PreDeploymentCommand = ""
		deploymentSettings.PostDeploymentCommand = ""
	}

	setting.MustSetData(deploymentSettings)
	return setting, nil
}

func (s *service) onCloneHttpSettingDefault(
	setting *entity.Setting,
	data *appCloneData,
) (*entity.Setting, error) {
	settings := data.CloneSettings
	routingSettings := setting.MustAsAppRoutingSettings()

	currDomains := routingSettings.Domains
	routingSettings.Domains = nil
	for _, copySettings := range settings.CloneHttpDomains {
		appDomain, _ := gofn.Find(currDomains, func(item *entity.AppDomain) bool {
			return item.Domain == copySettings.SourceDomain
		})
		if appDomain == nil {
			continue
		}
		appDomain.Domain = copySettings.TargetDomain
		appDomain.SSLCert = entity.ObjectID{ID: copySettings.TargetSSLCert.ID}
		// TODO: handle SSL cert validation
		routingSettings.Domains = append(routingSettings.Domains, appDomain)
	}

	setting.MustSetData(routingSettings)
	return setting, nil
}
