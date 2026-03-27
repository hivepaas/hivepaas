package entity

import (
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/base"
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
	ClientID       string         `json:"clientID"`
	ClientSecret   EncryptedField `json:"clientSecret"`
	Organization   string         `json:"org"`
	WebhookURL     string         `json:"webhookURL"`
	WebhookSecret  string         `json:"webhookSecret"`
	AppID          int64          `json:"appID"`
	InstallationID int64          `json:"installationID"`
	PrivateKey     EncryptedField `json:"privateKey"`
	SSOEnabled     bool           `json:"ssoEnabled"`
}

func (s *GithubApp) GetType() base.SettingType {
	return base.SettingTypeGithubApp
}

func (s *GithubApp) GetRefObjectIDs() *RefObjectIDs {
	return &RefObjectIDs{}
}

func (s *GithubApp) MustDecrypt() *GithubApp {
	s.ClientSecret.MustGetPlain()
	s.PrivateKey.MustGetPlain()
	return s
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

func (s *GithubApp) Migrate(setting *Setting) (hasChange bool, err error) {
	if CurrentGithubAppVersion == setting.Version {
		return false, nil
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentGithubAppVersion
	setting.MustSetData(s)
	return true, nil
}

func (s *Setting) AsGithubApp() (*GithubApp, error) {
	return parseSettingAs[*GithubApp](s)
}

func (s *Setting) MustAsGithubApp() *GithubApp {
	return gofn.Must(s.AsGithubApp())
}
