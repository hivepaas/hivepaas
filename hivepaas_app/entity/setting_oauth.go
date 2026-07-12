package entity

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

const (
	CurrentOAuthVersion = 1
)

var _ = registerSettingParser(base.SettingTypeOAuth, &oauthParser{})

type oauthParser struct {
}

func (s *oauthParser) New() SettingData {
	return &OAuth{}
}

type OAuth struct {
	ClientID         string         `json:"clientId"`
	ClientSecret     EncryptedField `json:"clientSecret"`
	Organization     string         `json:"org,omitempty"`
	AuthURL          string         `json:"authURL,omitempty"`
	TokenURL         string         `json:"tokenURL,omitempty"`
	ProfileURL       string         `json:"profileURL,omitempty"`
	AutoDiscoveryURL string         `json:"autoDiscoveryURL,omitempty"`
	Scopes           []string       `json:"scopes,omitempty"`
}

func (s *OAuth) GetType() base.SettingType {
	return base.SettingTypeOAuth
}

func (s *OAuth) GetRefObjectIDs() *RefObjectIDs {
	return &RefObjectIDs{}
}

func (s *OAuth) CalcResLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().CalcResLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *OAuth) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentOAuthVersion {
		return false, nil
	}
	if setting.Version > CurrentOAuthVersion {
		return false, apperrors.Wrap(apperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentOAuthVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}

func (s *OAuth) Decrypt() error {
	_, err := s.ClientSecret.GetPlain()
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (s *Setting) AsOAuth() (*OAuth, error) {
	// Github-app setting can be parsed as OAuth
	if s.Type == base.SettingTypeGithubApp {
		ghApp, err := s.AsGithubApp()
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		return ghApp.ConvertAsOAuth(), nil
	}
	return parseSettingAs[*OAuth](s)
}

func (s *Setting) MustAsOAuth() *OAuth {
	return gofn.Must(s.AsOAuth())
}
