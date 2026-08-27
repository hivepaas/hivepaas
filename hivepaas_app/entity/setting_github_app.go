package entity

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

const (
	CurrentGithubAppVersion = 1
)

var _ = registerSettingParser(base.SettingTypeGithubApp, &githubAppParser{})

type githubAppParser struct {
}

func (s *githubAppParser) New() SettingData {
	return &GithubApp{}
}

type GithubApp struct {
	ClientID       string         `json:"clientId"`
	ClientSecret   EncryptedField `json:"clientSecret"`
	Organization   string         `json:"org"`
	WebhookURL     string         `json:"webhookURL"`
	WebhookSecret  string         `json:"webhookSecret"` // NOTE: don't encrypt this, it's used in queries
	AppID          int64          `json:"appId"`
	InstallationID int64          `json:"installationId"`
	PrivateKey     EncryptedField `json:"privateKey"`
	SSOEnabled     bool           `json:"ssoEnabled"`
}

func (s *GithubApp) GetType() base.SettingType {
	return base.SettingTypeGithubApp
}

func (s *GithubApp) GetRefObjectIDs() *RefObjectIDs {
	return &RefObjectIDs{}
}

func (s *GithubApp) GetResourceLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().GetResourceLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *GithubApp) Decrypt() error {
	_, err := s.ClientSecret.GetPlain()
	if err != nil {
		return hperrors.Wrap(err)
	}
	_, err = s.PrivateKey.GetPlain()
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (s *GithubApp) ConvertAsOAuth() *OAuth {
	if !s.SSOEnabled {
		return nil
	}
	return &OAuth{
		ClientID:     s.ClientID,
		ClientSecret: s.ClientSecret,
		Organization: s.Organization,
	}
}

func (s *GithubApp) ConvertAsRepoWebhook() *RepoWebhook {
	return &RepoWebhook{
		Kind:   base.WebhookKindGithub,
		Secret: s.WebhookSecret,
	}
}

func (s *Setting) AsGithubApp() (*GithubApp, error) {
	return parseSettingAs[*GithubApp](s)
}

func (s *Setting) MustAsGithubApp() *GithubApp {
	return gofn.Must(s.AsGithubApp())
}
