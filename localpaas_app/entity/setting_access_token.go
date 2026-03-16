package entity

import (
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/base"
)

const (
	CurrentAccessTokenVersion = 1
)

var _ = registerSettingParser(base.SettingTypeAccessToken, &accessTokenParser{})

type accessTokenParser struct {
}

func (s *accessTokenParser) New() SettingData {
	return &AccessToken{}
}

type AccessToken struct {
	User    string         `json:"user"`
	Token   EncryptedField `json:"token"`
	BaseURL string         `json:"baseURL"`
}

func (s *AccessToken) GetType() base.SettingType {
	return base.SettingTypeAccessToken
}

func (s *AccessToken) GetRefObjectIDs() *RefObjectIDs {
	return &RefObjectIDs{}
}

func (s *AccessToken) MustDecrypt() *AccessToken {
	s.Token.MustGetPlain()
	return s
}

func (s *AccessToken) Migrate(setting *Setting) (hasChange bool, err error) {
	if CurrentAccessTokenVersion == setting.Version {
		return false, nil
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentAccessTokenVersion
	setting.MustSetData(s)
	return true, nil
}

func (s *Setting) AsAccessToken() (*AccessToken, error) {
	return parseSettingAs[*AccessToken](s)
}

func (s *Setting) MustAsAccessToken() *AccessToken {
	return gofn.Must(s.AsAccessToken())
}
