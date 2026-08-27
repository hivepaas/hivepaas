package appsettingsuc

import (
	"context"
	"strings"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
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
		return nil, hperrors.Wrap(err)
	}

	settings, _, err := uc.settingRepo.List(ctx, uc.db, nil, nil,
		bunex.SelectWhereIn("setting.type IN (?)", base.SettingTypeAppClone, base.SettingTypeAppRouting),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
		bunex.SelectWhere("setting.object_id = ?", app.ID),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	app.Settings = settings

	input := &appsettingsdto.AppCloneSettingsTransformInput{
		App: app,
	}

	err = uc.initDefaultAppCloneSettings(ctx, uc.db, input)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	err = uc.loadAppCloneSettingsRefData(ctx, uc.db, input)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	resp, err := appsettingsdto.TransformAppCloneSettings(input)
	if err != nil {
		return nil, hperrors.Wrap(err)
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
			CloneRoutingSettings:    true,

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

	routingSetting := app.GetSettingByType(base.SettingTypeAppRouting)
	refObjects := entity.NewRefObjects()
	var routingSettings *entity.AppRoutingSettings
	if routingSetting != nil {
		routingSettings = routingSetting.MustAsAppRoutingSettings()
		err = uc.settingService.LoadRefObjectsSkipMissing(ctx, db, &refObjects, app.GetObjectScope(),
			true, routingSetting)
		if err != nil {
			return hperrors.Wrap(err)
		}
	} else {
		routingSettings = &entity.AppRoutingSettings{}
	}

	currDomainSettings := cloneSettings.CloneRoutingDomains
	cloneSettings.CloneRoutingDomains = nil
	for _, domain := range routingSettings.Domains {
		domainSettings, _ := gofn.Find(currDomainSettings, func(d *entity.AppCloneRoutingDomainSettings) bool {
			return d.SourceDomain == domain.Domain
		})
		if domainSettings == nil {
			domainSettings = &entity.AppCloneRoutingDomainSettings{
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

		cloneSettings.CloneRoutingDomains = append(cloneSettings.CloneRoutingDomains, domainSettings)
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
		return hperrors.Wrap(err)
	}
	for _, settingID := range refIDs.RefSettingIDs {
		setting := input.RefObjects.RefSettings[settingID]
		if setting != nil {
			setting.CurrentObjectID = app.ID
		}
	}

	return nil
}
