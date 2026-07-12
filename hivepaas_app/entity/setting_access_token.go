package entity

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
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

func (s *AccessToken) CalcResLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().CalcResLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *AccessToken) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentAccessTokenVersion {
		return false, nil
	}
	if setting.Version > CurrentAccessTokenVersion {
		return false, apperrors.Wrap(apperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentAccessTokenVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}

func (s *AccessToken) Decrypt() error {
	_, err := s.Token.GetPlain()
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (s *Setting) AsAccessToken() (*AccessToken, error) {
	return parseSettingAs[*AccessToken](s)
}

func (s *Setting) MustAsAccessToken() *AccessToken {
	return gofn.Must(s.AsAccessToken())
}
