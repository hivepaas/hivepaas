package githubappuc

import (
	"context"

	gogithub "github.com/google/go-github/v85/github"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/reflectutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/githubappuc/githubappdto"
	"github.com/hivepaas/hivepaas/services/git/github"
)

func (uc *UC) CreateGithubApp(
	ctx context.Context,
	auth *basedto.Auth,
	req *githubappdto.CreateGithubAppReq,
) (*githubappdto.CreateGithubAppResp, error) {
	req.Type = currentSettingType
	githubApp := req.ToEntity()
	resp, err := uc.CreateSetting(ctx, &req.CreateSettingReq, &settings.CreateSettingData{
		VerifyingName:   gofn.Coalesce(req.Name, req.Organization),
		VerifyingRefIDs: githubApp.GetRefObjectIDs(),
		Version:         currentSettingVersion,
		PrepareCreation: func(
			ctx context.Context,
			db database.Tx,
			data *settings.CreateSettingData,
			pData *settings.PersistingSettingCreationData,
		) error {
			pData.Setting.Kind = string(base.SettingTypeGithubApp)
			err := uc.installGithubAppWebhook(ctx, pData.Setting.ID, githubApp, false)
			if err != nil {
				return apperrors.New(err)
			}
			err = pData.Setting.SetData(githubApp)
			if err != nil {
				return apperrors.New(err)
			}
			return nil
		},
	})
	if err != nil {
		return nil, apperrors.New(err)
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
	settingID string,
	githubApp *entity.GithubApp,
	update bool,
) error {
	if !update {
		githubApp.WebhookSecret = gofn.RandTokenAsHex(base.DefaultWebhookSecretByteLen)
	}

	if config.Current.IsDevEnv() && config.Current.Platform == config.PlatformLocal {
		githubApp.WebhookSecret = "abc123"
		githubApp.WebhookURL = "https://smee.io/RBNiNjxieUIWZ6Ej"
	} else {
		githubApp.WebhookURL = config.Current.RepoWebhookURL(settingID)
	}

	privateKey, err := githubApp.PrivateKey.GetPlain()
	if err != nil {
		return apperrors.New(err)
	}

	client, err := github.NewFromApp(githubApp.AppID, githubApp.InstallationID,
		reflectutil.UnsafeStrToBytes(privateKey))
	if err != nil {
		return apperrors.New(err)
	}

	shouldSet := true
	if !update {
		hook, err := client.GetAppHookConfig(ctx)
		if err != nil {
			return apperrors.New(err)
		}
		shouldSet = gofn.PtrValueOrEmpty(hook.URL) != githubApp.WebhookURL
	}

	if shouldSet {
		err = client.UpdateAppHookConfig(ctx, func(opts *gogithub.HookConfig) {
			opts.ContentType = new("json")
			opts.URL = &githubApp.WebhookURL
		})
		if err != nil {
			return apperrors.New(err)
		}
	}
	return nil
}
