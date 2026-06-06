package github

import (
	"context"

	gogithub "github.com/google/go-github/v85/github"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/infra/httpclient"
)

// AppManifest
// nolint https://docs.github.com/en/apps/sharing-github-apps/registering-a-github-app-from-a-manifest#github-app-manifest-parameters
type AppManifest struct {
	Name                  string            `json:"name,omitempty"`
	URL                   string            `json:"url"`
	Hook                  *AppManifestHook  `json:"hook_attributes"`
	RedirectURL           string            `json:"redirect_url"`
	CallbackURLs          []string          `json:"callback_urls"`
	SetupURL              string            `json:"setup_url"`
	Description           string            `json:"description,omitempty"`
	Public                bool              `json:"public,omitempty"`
	DefaultEvents         []string          `json:"default_events"`
	DefaultPermissions    map[string]string `json:"default_permissions"`
	RequestOAuthOnInstall bool              `json:"request_oauth_on_install,omitempty"`
	SetupOnUpdate         bool              `json:"setup_on_update,omitempty"`
}

type AppManifestHook struct {
	URL    string `json:"url"`
	Active bool   `json:"active"`
}

func AppManifestFlowComplete(
	ctx context.Context,
	code string,
) (*gogithub.AppConfig, error) {
	client := gogithub.NewClient(httpclient.DefaultClient)
	appConfig, _, err := client.Apps.CompleteAppManifest(ctx, code)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	return appConfig, nil
}
