package githubappuc

import (
	"context"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/config"
	"github.com/localpaas/localpaas/localpaas_app/entity/cacheentity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/githubappuc/githubappdto"
	"github.com/localpaas/localpaas/services/git/github"
)

func (uc *UC) BeginReprovisionGithubApp(
	ctx context.Context,
	auth *basedto.Auth,
	req *githubappdto.BeginReprovisionGithubAppReq,
) (*githubappdto.BeginReprovisionGithubAppResp, error) {
	appSetting, err := uc.GetSettingByID(ctx, uc.DB, &req.BaseSettingReq, req.ID, true)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	if appSetting.UpdateVer != req.UpdateVer {
		return nil, apperrors.Wrap(apperrors.ErrUpdateVerMismatched)
	}

	cfg := config.Current
	isLocalEnv := cfg.IsDevEnv() && cfg.Platform == config.PlatformLocal
	timeNow := timeutil.NowUTC()

	appSetting.UpdatedAt = timeNow
	appSetting.Name = req.Name

	githubApp := appSetting.MustAsGithubApp()
	if isLocalEnv {
		githubApp.WebhookSecret = webhookSecretLocal
		githubApp.WebhookURL = webhookURLLocal
	} else {
		githubApp.WebhookSecret = gofn.RandTokenAsHex(webhookSecretLen)
		githubApp.WebhookURL = cfg.RepoWebhookURL(base.WebhookKindGithub, githubApp.WebhookURL)
	}
	appSetting.MustSetData(githubApp)

	state := gofn.RandTokenAsHex(appManifestStateLen)
	manifest := &github.AppManifest{
		Name:         appSetting.Name,
		URL:          cfg.BaseURL,
		CallbackURLs: []string{cfg.SsoCallbackURL(appSetting.ID)},
		Hook: &github.AppManifestHook{
			URL:    githubApp.WebhookURL,
			Active: true,
		},
		Public:             false,
		DefaultEvents:      defaultAppEvents,
		DefaultPermissions: defaultAppPermissions,
		SetupOnUpdate:      false,
	}

	var beginFlowURL string
	switch req.Scope.ScopeType() { //nolint:exhaustive
	case base.SettingScopeGlobal:
		beginFlowURL = cfg.GlobalGithubAppManifestFlowBeginURL(appSetting.ID, state)
		manifest.RedirectURL = cfg.GlobalGithubAppManifestFlowProgressURL(appSetting.ID)
		manifest.SetupURL = manifest.RedirectURL
	case base.SettingScopeProject:
		beginFlowURL = cfg.ProjectGithubAppManifestFlowBeginURL(req.Scope.ProjectID, appSetting.ID, state)
		manifest.RedirectURL = cfg.ProjectGithubAppManifestFlowProgressURL(req.Scope.ProjectID, appSetting.ID)
		manifest.SetupURL = manifest.RedirectURL
	default:
		return nil, apperrors.New(apperrors.ErrUnsupported)
	}

	manifestCache := &cacheentity.GithubAppManifest{
		Manifest:    manifest,
		State:       state,
		Reprovision: true,
		GithubApp:   appSetting,
	}

	err = uc.cacheAppManifestRepo.Set(ctx, appSetting.ID, manifestCache, appManifestCacheExp)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &githubappdto.BeginReprovisionGithubAppResp{
		Data: &githubappdto.BeginReprovisionGithubAppDataResp{
			RedirectURL: beginFlowURL,
		},
	}, nil
}
