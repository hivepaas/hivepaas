package appsettingsuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/settinghelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
)

func (uc *UC) GetAppDeploymentSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *appsettingsdto.GetAppDeploymentSettingsReq,
) (*appsettingsdto.GetAppDeploymentSettingsResp, error) {
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
		bunex.SelectWhere("setting.type = ?", base.SettingTypeAppDeployment),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
		bunex.SelectWhere("setting.object_id = ?", app.ID), // load app direct settings
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	input := &appsettingsdto.AppDeploymentSettingsTransformInput{
		App:                app,
		DeploymentSettings: settinghelper.FindSettingByType(settings, base.SettingTypeAppDeployment),
	}
	err = uc.loadAppDeploymentSettingsRefData(ctx, uc.db, input)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	resp, err := appsettingsdto.TransformDeploymentSettings(input)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &appsettingsdto.GetAppDeploymentSettingsResp{
		Data: resp,
	}, nil
}

func (uc *UC) loadAppDeploymentSettingsRefData(
	ctx context.Context,
	db database.IDB,
	input *appsettingsdto.AppDeploymentSettingsTransformInput,
) (err error) {
	app := input.App
	service, err := uc.clusterService.ServiceInspect(ctx, app.ServiceID, true)
	if err != nil {
		return hperrors.Wrap(err)
	}
	input.ServiceSpec = &service.Spec

	refIDs := &entity.RefObjectIDs{}
	if input.DeploymentSettings != nil {
		refIDs = input.DeploymentSettings.MustAsAppDeploymentSettings().GetRefObjectIDs()
	}

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
