package githubappuc

import (
	"context"
	"time"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/config"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/entity/cacheentity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/ulid"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/githubappuc/githubappdto"
	"github.com/localpaas/localpaas/services/git/github"
)

const (
	appManifestStateLen = 24
	appManifestCacheExp = 10 * time.Minute
)

var (
	defaultAppEvents = []string{
		"push",
		// "create",
	}

	defaultAppPermissions = map[string]string{
		"contents": "read",
		// "repository_hooks": "write",
		// "organization_hooks": "write",
		// "repository_projects": "read",
		// "pull_requests": "read",
		// "organization_personal_access_tokens": "read",
	}

	webhookSecretLocal = "abc123"
	webhookURLLocal    = "https://smee.io/RBNiNjxieUIWZ6Ej"
)

func (uc *UC) BeginGithubAppManifestFlow(
	ctx context.Context,
	auth *basedto.Auth,
	req *githubappdto.BeginGithubAppManifestFlowReq,
) (*githubappdto.BeginGithubAppManifestFlowResp, error) {
	cfg := config.Current
	isLocalEnv := cfg.IsDevEnv() && cfg.Platform == config.PlatformLocal
	timeNow := timeutil.NowUTC()

	appSetting := &entity.Setting{
		ID:              gofn.Must(ulid.NewStringULID()),
		Scope:           req.Scope.ScopeType(),
		ObjectID:        req.Scope.MainObjectID(),
		Type:            base.SettingTypeGithubApp,
		Kind:            string(base.SettingTypeGithubApp),
		Status:          base.SettingStatusActive,
		Name:            gofn.Coalesce(req.Name, "my localpaas app"),
		AvailInProjects: req.AvailInProjects,
		Default:         req.Default,
		Version:         entity.CurrentGithubAppVersion,
		CreatedAt:       timeNow,
		UpdatedAt:       timeNow,
	}
	githubApp := &entity.GithubApp{
		Organization: req.Org,
		SSOEnabled:   req.SSOEnabled,
	}
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
		Manifest:  manifest,
		State:     state,
		GithubApp: appSetting,
	}

	err := uc.cacheAppManifestRepo.Set(ctx, appSetting.ID, manifestCache, appManifestCacheExp)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &githubappdto.BeginGithubAppManifestFlowResp{
		Data: &githubappdto.BeginGithubAppManifestFlowDataResp{
			RedirectURL: beginFlowURL,
			SettingID:   appSetting.ID,
		},
	}, nil
}
