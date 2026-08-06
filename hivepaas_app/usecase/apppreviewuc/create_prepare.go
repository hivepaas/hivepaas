package apppreviewuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/apppreviewuc/apppreviewdto"
)

func (uc *UC) PrepareCreatePreview(
	ctx context.Context,
	auth *basedto.Auth,
	req *apppreviewdto.PrepareCreatePreviewReq,
) (_ *apppreviewdto.PrepareCreatePreviewResp, err error) {
	app, featureSettings, err := uc.appService.LoadAppWithFeatureSettings(ctx, uc.db, req.ProjectID, req.AppID,
		true, true,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
		bunex.SelectRelation("ProjectEnv"),
		bunex.SelectRelation("Settings",
			bunex.SelectWhere("setting.type = ?", base.SettingTypeAppDeployment),
		),
	)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp := &apppreviewdto.PrepareCreatePreviewResp{
		Data: &apppreviewdto.PrepareCreatePreviewDataResp{Enabled: true},
	}
	previewSettings := featureSettings.PreviewSettings
	if previewSettings != nil {
		if !previewSettings.Enabled {
			resp.Data.Enabled = false
			return resp, nil
		}
		hasAppsToClone := len(previewSettings.AppsToClone) > 0
		if hasAppsToClone && previewSettings.AutoCloneApps {
			resp.Data.CanSkipCloningDBApps = true
		} else if hasAppsToClone && !previewSettings.AutoCloneApps {
			resp.Data.CanCloneDBApps = true
		}
	}

	deploymentSetting := app.GetSettingByType(base.SettingTypeAppDeployment)
	if deploymentSetting == nil {
		return nil, apperrors.NewNotFound("Deployment settings")
	}
	deploymentSettings := deploymentSetting.MustAsAppDeploymentSettings()
	repoSource := deploymentSettings.RepoSource
	if deploymentSettings.ActiveMethod != base.DeploymentMethodRepo || repoSource == nil {
		return nil, apperrors.Wrap(apperrors.ErrDeploymentMethodRepoRequired)
	}

	refObjects := entity.NewRefObjects()
	err = uc.settingService.LoadRefObjects(ctx, uc.db, &refObjects, app.GetObjectScope(),
		true, true, app.Settings...)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	respData := &apppreviewdto.PrepareCreatePreviewDataResp{
		RepoURL: repoSource.RepoURL,
	}
	if repoSource.Credentials.ID != "" {
		respData.RepoCredentials = &basedto.ObjectIDResp{ID: repoSource.Credentials.ID}
	}

	credSetting := refObjects.RefSettings[repoSource.Credentials.ID]
	if credSetting != nil {
		if credSetting.Type == base.SettingTypeGithubApp || credSetting.Type == base.SettingTypeAccessToken {
			respData.CanListBranches = true
			respData.CanListPullRequests = true
		}
	}

	return &apppreviewdto.PrepareCreatePreviewResp{
		Data: respData,
	}, nil
}
