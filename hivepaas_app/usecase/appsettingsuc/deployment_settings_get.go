package appsettingsuc

import (
	"context"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
)

func (uc *UC) GetAppDeploymentSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *appsettingsdto.GetAppDeploymentSettingsReq,
) (*appsettingsdto.GetAppDeploymentSettingsResp, error) {
	app, err := uc.appRepo.GetByID(ctx, uc.db, req.ProjectID, req.AppID,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
	)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	settings, _, err := uc.settingRepo.List(ctx, uc.db, nil, nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeAppDeployment),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
		bunex.SelectWhere("setting.object_id = ?", app.ID), // load app direct settings
	)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	input := &appsettingsdto.AppDeploymentSettingsTransformInput{
		App:                app,
		DeploymentSettings: gofn.FirstOr(settings, nil),
	}
	err = uc.loadAppDeploymentSettingsRefData(ctx, uc.db, input)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp, err := appsettingsdto.TransformDeploymentSettings(input)
	if err != nil {
		return nil, apperrors.Wrap(err)
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
		return apperrors.Wrap(err)
	}
	input.ServiceSpec = &service.Spec

	refIDs := &entity.RefObjectIDs{}
	if input.DeploymentSettings != nil {
		refIDs = input.DeploymentSettings.MustAsAppDeploymentSettings().GetRefObjectIDs()
	}

	refObjects, err := uc.settingService.LoadReferenceObjectsByIDs(ctx, db, app.GetObjectScope(),
		true, false, refIDs)
	if err != nil {
		return apperrors.Wrap(err)
	}
	for _, setting := range refObjects.RefSettings {
		setting.CurrentObjectID = app.ID
	}
	input.RefObjects = refObjects

	return nil
}
