package githubappuc

import (
	"context"

	gogithub "github.com/google/go-github/v79/github"
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/config"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/reflectutil"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/githubappuc/githubappdto"
	"github.com/localpaas/localpaas/services/git/github"
)

const (
	currentSettingType    = base.SettingTypeGithubApp
	currentSettingVersion = entity.CurrentGithubAppVersion

	webhookSecretLen = 24
)

func (uc *UC) CreateGithubApp(
	ctx context.Context,
	auth *basedto.Auth,
	req *githubappdto.CreateGithubAppReq,
) (*githubappdto.CreateGithubAppResp, error) {
	req.Type = currentSettingType
	resp, err := uc.CreateSetting(ctx, &req.CreateSettingReq, &settings.CreateSettingData{
		VerifyingName: gofn.Coalesce(req.Name, req.Organization),
		Version:       currentSettingVersion,
		PrepareCreation: func(
			ctx context.Context,
			db database.Tx,
			data *settings.CreateSettingData,
			pData *settings.PersistingSettingCreationData,
		) error {
			pData.Setting.Kind = string(base.SettingTypeGithubApp)
			githubApp := req.ToEntity()
			err := uc.installGithubAppWebhook(ctx, githubApp, false)
			if err != nil {
				return apperrors.Wrap(err)
			}
			err = pData.Setting.SetData(githubApp)
			if err != nil {
				return apperrors.Wrap(err)
			}
			return nil
		},
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &githubappdto.CreateGithubAppResp{
		Data: &githubappdto.GithubAppCreationResp{
			ID:          resp.Data.ID,
			CallbackURL: config.Current.SsoCallbackURL(resp.Data.ID),
		},
	}, nil
}

func (uc *UC) installGithubAppWebhook(
	ctx context.Context,
	githubApp *entity.GithubApp,
	update bool,
) error {
	if !update {
		githubApp.WebhookSecret = gofn.RandTokenAsHex(webhookSecretLen)
	}

	if config.Current.IsDevEnv() && config.Current.Platform == config.PlatformLocal {
		githubApp.WebhookSecret = "abc123"
		githubApp.WebhookURL = "https://smee.io/RBNiNjxieUIWZ6Ej"
	} else {
		githubApp.WebhookURL = config.Current.RepoWebhookURL(base.WebhookKindGithub, githubApp.WebhookSecret)
	}

	client, err := github.NewFromApp(githubApp.AppID, githubApp.InstallationID,
		reflectutil.UnsafeStrToBytes(githubApp.PrivateKey.MustGetPlain()))
	if err != nil {
		return apperrors.Wrap(err)
	}

	shouldSet := true
	if !update {
		hook, err := client.GetAppHookConfig(ctx)
		if err != nil {
			return apperrors.Wrap(err)
		}
		shouldSet = gofn.PtrValueOrEmpty(hook.URL) != githubApp.WebhookURL
	}

	if shouldSet {
		err = client.UpdateAppHookConfig(ctx, func(opts *gogithub.HookConfig) {
			opts.ContentType = new("json")
			opts.URL = &githubApp.WebhookURL
		})
		if err != nil {
			return apperrors.Wrap(err)
		}
	}
	return nil
}
