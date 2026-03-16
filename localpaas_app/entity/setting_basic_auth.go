package entity

import (
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/base"
)

const (
	CurrentBasicAuthVersion = 1
)

var _ = registerSettingParser(base.SettingTypeBasicAuth, &basicAuthParser{})

type basicAuthParser struct {
}

func (s *basicAuthParser) New() SettingData {
	return &BasicAuth{}
}

type BasicAuth struct {
	Username string         `json:"username"`
	Password EncryptedField `json:"password"`
}

func (s *BasicAuth) GetType() base.SettingType {
	return base.SettingTypeBasicAuth
}

func (s *BasicAuth) GetRefObjectIDs() *RefObjectIDs {
	return &RefObjectIDs{}
}

func (s *BasicAuth) MustDecrypt() *BasicAuth {
	s.Password.MustGetPlain()
	return s
}

func (s *BasicAuth) Migrate(setting *Setting) (hasChange bool, err error) {
	if CurrentBasicAuthVersion == setting.Version {
		return false, nil
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentBasicAuthVersion
	setting.MustSetData(s)
	return true, nil
}

func (s *Setting) AsBasicAuth() (*BasicAuth, error) {
	return parseSettingAs[*BasicAuth](s)
}

func (s *Setting) MustAsBasicAuth() *BasicAuth {
	return gofn.Must(s.AsBasicAuth())
}
