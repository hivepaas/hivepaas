package appuc

import (
	"context"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/usecase/appuc/appdto"
)

func (uc *UC) GetAppDeploymentSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *appdto.GetAppDeploymentSettingsReq,
) (*appdto.GetAppDeploymentSettingsResp, error) {
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

	input := &appdto.AppDeploymentSettingsTransformInput{
		App:                app,
		DeploymentSettings: gofn.FirstOr(settings, nil),
	}
	err = uc.loadAppDeploymentSettingsRefData(ctx, uc.db, input)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp, err := appdto.TransformDeploymentSettings(input)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &appdto.GetAppDeploymentSettingsResp{
		Data: resp,
	}, nil
}

func (uc *UC) loadAppDeploymentSettingsRefData(
	ctx context.Context,
	db database.IDB,
	input *appdto.AppDeploymentSettingsTransformInput,
) (err error) {
	app := input.App
	service, err := uc.appService.ServiceInspect(ctx, app.ServiceID, true)
	if err != nil {
		return apperrors.Wrap(err)
	}
	input.ServiceSpec = &service.Spec

	refIDs := &entity.RefObjectIDs{}
	if input.DeploymentSettings != nil {
		refIDs = input.DeploymentSettings.MustAsAppDeploymentSettings().GetRefObjectIDs()
	}

	refObjects, err := uc.settingService.LoadReferenceObjectsByIDs(ctx, db, app.GetSettingScope(),
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
