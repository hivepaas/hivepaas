package githubappuc

import (
	"context"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity/cacheentity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/githubappuc/githubappdto"
	"github.com/hivepaas/hivepaas/services/git/github"
)

const (
	appManifestStateLen = 24
	appManifestCacheExp = 10 * time.Minute
)

var (
	defaultAppEvents = []string{
		"push",
		"issue_comment",
		"pull_request",
		// "create",
	}

	defaultAppPermissions = map[string]string{
		"contents":      "read",
		"issues":        "write",
		"pull_requests": "write",
		// "repository_hooks": "write",
		// "organization_hooks": "write",
		// "repository_projects": "read",
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
		Name:            gofn.Coalesce(req.Name, "my hivepaas app"),
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
		githubApp.WebhookSecret = gofn.RandTokenAsHex(base.DefaultWebhookSecretByteLen)
		githubApp.WebhookURL = cfg.RepoWebhookURL(appSetting.ID)
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
	switch req.Scope.ScopeType() {
	case base.ObjectScopeGlobal:
		beginFlowURL = cfg.GlobalGithubAppManifestFlowBeginURL(appSetting.ID, state)
		manifest.RedirectURL = cfg.GlobalGithubAppManifestFlowProgressURL(appSetting.ID)
		manifest.SetupURL = manifest.RedirectURL
	case base.ObjectScopeProject:
		beginFlowURL = cfg.ProjectGithubAppManifestFlowBeginURL(req.Scope.ProjectID, appSetting.ID, state)
		manifest.RedirectURL = cfg.ProjectGithubAppManifestFlowProgressURL(req.Scope.ProjectID, appSetting.ID)
		manifest.SetupURL = manifest.RedirectURL
	case base.ObjectScopeApp, base.ObjectScopeUser:
		fallthrough
	default:
		return nil, apperrors.New(apperrors.ErrObjectScopeInvalid).
			WithParam("Scope", req.Scope.ScopeType())
	}

	manifestCache := &cacheentity.GithubAppManifest{
		Manifest:  manifest,
		State:     state,
		GithubApp: appSetting,
	}

	err := uc.cacheAppManifestRepo.Set(ctx, appSetting.ID, manifestCache, appManifestCacheExp)
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &githubappdto.BeginGithubAppManifestFlowResp{
		Data: &githubappdto.BeginGithubAppManifestFlowDataResp{
			RedirectURL: beginFlowURL,
			SettingID:   appSetting.ID,
		},
	}, nil
}
