package appfeaturesettingsuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/appfeaturesettingsuc/appfeaturesettingsdto"
)

func (uc *UC) UpdateAppFeatureSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *appfeaturesettingsdto.UpdateAppFeatureSettingsReq,
) (*appfeaturesettingsdto.UpdateAppFeatureSettingsResp, error) {
	req.Type = currentSettingType
	_, err := uc.UpdateUniqueSetting(ctx, &req.UpdateUniqueSettingReq, &settings.UpdateUniqueSettingData{
		Name: "App feature settings",
		PrepareUpdate: func(
			ctx context.Context,
			db database.Tx,
			data *settings.UpdateUniqueSettingData,
			pData *settings.PersistingSettingData,
		) error {
			featureSettings := req.ToEntity()
			if err := uc.validateAppsToClone(ctx, db, req.Scope, featureSettings); err != nil {
				return hperrors.Wrap(err)
			}
			if err := pData.Setting.SetData(featureSettings); err != nil {
				return hperrors.Wrap(err)
			}
			return nil
		},
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &appfeaturesettingsdto.UpdateAppFeatureSettingsResp{}, nil
}

func (uc *UC) validateAppsToClone(
	ctx context.Context,
	db database.IDB,
	scope *entity.ObjectScope,
	featureSettings *entity.AppFeatureSettings,
) error {
	if featureSettings.PreviewSettings == nil {
		return nil
	}
	appsToValidate := featureSettings.PreviewSettings.AppsToClone.ToIDStringSlice()
	if len(appsToValidate) == 0 {
		return nil
	}

	currApp := scope.App
	for _, appToValidate := range appsToValidate {
		if appToValidate == currApp.ID {
			return hperrors.Wrap(hperrors.ErrAppIsCurrent)
		}
	}

	// TODO (high): User must have WRITE on the current env?

	apps, err := uc.AppService.LoadApps(ctx, db, currApp.ProjectID, appsToValidate, true, true,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
		bunex.SelectRelation("ProjectEnv"),
		bunex.SelectRelation("Settings",
			bunex.SelectWhere("setting.type = ?", base.SettingTypeAppClone),
			bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
		),
	)
	if err != nil {
		return hperrors.Wrap(err)
	}

	for _, app := range apps {
		if app.ProjectEnvID != currApp.ProjectEnvID {
			return hperrors.Wrap(hperrors.ErrAppsNotInSameProjectEnv).WithMsgLog("app '%v'", app.ProjectEnvID)
		}
		cloneSettings := app.GetSettingByType(base.SettingTypeAppClone)
		if cloneSettings == nil {
			return hperrors.Wrap(hperrors.ErrAppCloneSettingsRequired).WithParam("Name", app.Name)
		}
	}

	return nil
}
