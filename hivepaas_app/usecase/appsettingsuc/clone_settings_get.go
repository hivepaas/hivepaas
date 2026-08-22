package appsettingsuc

import (
	"context"
	"strings"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
)

func (uc *UC) GetAppCloneSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *appsettingsdto.GetAppCloneSettingsReq,
) (*appsettingsdto.GetAppCloneSettingsResp, error) {
	app, err := uc.appService.LoadApp(ctx, uc.db, req.ProjectID, req.AppID, false, false,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
		bunex.SelectRelation("ProjectEnv"),
	)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	settings, _, err := uc.settingRepo.List(ctx, uc.db, nil, nil,
		bunex.SelectWhereIn("setting.type IN (?)", base.SettingTypeAppClone, base.SettingTypeAppRouting),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
		bunex.SelectWhere("setting.object_id = ?", app.ID),
	)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	app.Settings = settings

	input := &appsettingsdto.AppCloneSettingsTransformInput{
		App: app,
	}

	err = uc.initDefaultAppCloneSettings(ctx, uc.db, input)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	err = uc.loadAppCloneSettingsRefData(ctx, uc.db, input)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp, err := appsettingsdto.TransformAppCloneSettings(input)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &appsettingsdto.GetAppCloneSettingsResp{
		Data: resp,
	}, nil
}

func (uc *UC) initDefaultAppCloneSettings(
	ctx context.Context,
	db database.IDB,
	input *appsettingsdto.AppCloneSettingsTransformInput,
) (err error) {
	app := input.App
	cloneSetting := app.GetSettingByType(base.SettingTypeAppClone)
	var cloneSettings *entity.AppCloneSettings

	if cloneSetting != nil {
		input.UpdateVer = cloneSetting.UpdateVer
		cloneSettings = cloneSetting.MustAsAppCloneSettings()
	} else {
		cloneSettings = &entity.AppCloneSettings{
			TargetName:     app.Name + " (clone)",
			TargetEnv:      app.ProjectEnv.Name,
			TargetStatus:   base.AppStatusActive,
			TargetReplicas: -1,

			CloneDeploymentSettings: true,
			CloneHttpSettings:       true,

			CloneVolumes:    true,
			CloneVolumeData: true,
			LiveVolumeClone: true,

			CloneEnvVars:      true,
			CloneSecrets:      true,
			CloneConfigFiles:  true,
			ClonePeriodicJobs: true,
			CloneSchedJobs:    true,

			Notification: &entity.BaseEventNotification{
				SuccessUseDefault: true,
				FailureUseDefault: true,
			},
		}
	}

	httpSetting := app.GetSettingByType(base.SettingTypeAppRouting)
	refObjects := entity.NewRefObjects()
	var httpSettings *entity.AppRoutingSettings
	if httpSetting != nil {
		httpSettings = httpSetting.MustAsAppRoutingSettings()
		err = uc.settingService.LoadRefObjectsSkipMissing(ctx, db, &refObjects, app.GetObjectScope(),
			true, httpSetting)
		if err != nil {
			return apperrors.Wrap(err)
		}
	} else {
		httpSettings = &entity.AppRoutingSettings{}
	}

	currDomainSettings := cloneSettings.CloneHttpDomains
	cloneSettings.CloneHttpDomains = nil
	for _, domain := range httpSettings.Domains {
		domainSettings, _ := gofn.Find(currDomainSettings, func(d *entity.AppCloneHttpDomainSettings) bool {
			return d.SourceDomain == domain.Domain
		})
		if domainSettings == nil {
			domainSettings = &entity.AppCloneHttpDomainSettings{
				SourceDomain: domain.Domain,
				TargetDomain: "clone_" + domain.Domain,
			}
		} else {
			domainSettings.TargetSSLCert.ID = ""
		}
		domainSettings.SourceSSLCert = domain.SSLCert

		sslCert := refObjects.RefSettings[domain.SSLCert.ID]
		if sslCert != nil && strings.HasPrefix(sslCert.MustAsSSLCert().Domain, "*.") {
			domainSettings.TargetSSLCert = domain.SSLCert
		}

		cloneSettings.CloneHttpDomains = append(cloneSettings.CloneHttpDomains, domainSettings)
	}

	input.AppCloneSettings = cloneSettings
	return nil
}

func (uc *UC) loadAppCloneSettingsRefData(
	ctx context.Context,
	db database.IDB,
	input *appsettingsdto.AppCloneSettingsTransformInput,
) (err error) {
	if input.AppCloneSettings == nil {
		return nil
	}

	app := input.App
	refIDs := input.AppCloneSettings.GetRefObjectIDs()
	err = uc.settingService.LoadRefObjectsByIDsSkipMissing(ctx, db, &input.RefObjects, app.GetObjectScope(),
		true, refIDs)
	if err != nil {
		return apperrors.Wrap(err)
	}
	for _, settingID := range refIDs.RefSettingIDs {
		setting := input.RefObjects.RefSettings[settingID]
		if setting != nil {
			setting.CurrentObjectID = app.ID
		}
	}

	return nil
}
