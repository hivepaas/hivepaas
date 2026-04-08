package githubappuc

import (
	"context"
	"fmt"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/config"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/githubappuc/githubappdto"
	"github.com/localpaas/localpaas/services/git/github"
)

func (uc *UC) HandleGithubAppManifestFlowProgress(
	ctx context.Context,
	req *githubappdto.HandleGithubAppManifestFlowProgressReq,
) (*githubappdto.HandleGithubAppManifestFlowProgressResp, error) {
	if req.InstallationID == 0 && req.Code != "" && req.State != "" {
		return uc.handleGithubAppManifestFlowOnCreation(ctx, req)
	}

	if req.InstallationID > 0 {
		return uc.handleGithubAppManifestFlowOnInstallation(ctx, req)
	}

	return nil, apperrors.NewNotImplemented()
}

func (uc *UC) handleGithubAppManifestFlowOnCreation(
	ctx context.Context,
	req *githubappdto.HandleGithubAppManifestFlowProgressReq,
) (*githubappdto.HandleGithubAppManifestFlowProgressResp, error) {
	manifestCache, err := uc.cacheAppManifestRepo.Get(ctx, req.SettingID)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	if req.State != manifestCache.State {
		return nil, apperrors.NewParamInvalid("State").WithMsgLog("param 'state' must match")
	}

	appConfig, err := github.AppManifestFlowComplete(ctx, req.Code)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	appSetting := manifestCache.GithubApp
	appSetting.Name = *appConfig.Name
	githubApp := appSetting.MustAsGithubApp()
	githubApp.AppID = gofn.PtrValueOrEmpty(appConfig.ID)
	githubApp.ClientID = gofn.PtrValueOrEmpty(appConfig.ClientID)
	githubApp.ClientSecret = entity.NewEncryptedField(gofn.PtrValueOrEmpty(appConfig.ClientSecret))
	githubApp.PrivateKey = entity.NewEncryptedField(gofn.PtrValueOrEmpty(appConfig.PEM))
	appSetting.MustSetData(githubApp)

	err = uc.cacheAppManifestRepo.Set(ctx, appSetting.ID, manifestCache, appManifestCacheExp)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	var redirectURL string
	// Redirect the user to the page where they can install the newly created app
	if githubApp.Organization != "" {
		redirectURL = fmt.Sprintf("https://github.com/organizations/%s/settings/apps/%s/installations",
			githubApp.Organization, *appConfig.Slug)
	} else {
		redirectURL = fmt.Sprintf("https://github.com/settings/apps/%s/installations",
			*appConfig.Slug)
	}

	return &githubappdto.HandleGithubAppManifestFlowProgressResp{
		Data: &githubappdto.HandleGithubAppManifestFlowProgressDataResp{
			RedirectURL: redirectURL,
		},
	}, nil
}

func (uc *UC) handleGithubAppManifestFlowOnInstallation(
	ctx context.Context,
	req *githubappdto.HandleGithubAppManifestFlowProgressReq,
) (*githubappdto.HandleGithubAppManifestFlowProgressResp, error) {
	manifestCache, err := uc.cacheAppManifestRepo.Get(ctx, req.SettingID)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	// TODO: query installation list of the app and validate this one

	appSetting := manifestCache.GithubApp
	githubApp := appSetting.MustAsGithubApp()
	githubApp.InstallationID = req.InstallationID
	appSetting.MustSetData(githubApp)

	if manifestCache.Reprovision { //nolint:nestif
		// Loads the setting from DB to check update version matching
		dbSetting, err := uc.SettingRepo.GetByID(ctx, uc.DB, nil, "", appSetting.ID, false)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		if dbSetting.UpdateVer != appSetting.UpdateVer {
			return nil, apperrors.Wrap(apperrors.ErrUpdateVerMismatched)
		}

		appSetting.UpdateVer++
		err = uc.SettingRepo.Update(ctx, uc.DB, appSetting)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
	} else {
		err = uc.SettingRepo.Insert(ctx, uc.DB, appSetting)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
	}

	_ = uc.cacheAppManifestRepo.Del(ctx, req.SettingID)

	var redirectURL string
	if appSetting.ObjectID == "" {
		redirectURL = config.Current.DashboardGlobalGithubAppsURL()
	} else {
		redirectURL = config.Current.DashboardProjectGithubAppsURL(appSetting.ObjectID)
	}

	return &githubappdto.HandleGithubAppManifestFlowProgressResp{
		Data: &githubappdto.HandleGithubAppManifestFlowProgressDataResp{
			RedirectURL: redirectURL,
		},
	}, nil
}
