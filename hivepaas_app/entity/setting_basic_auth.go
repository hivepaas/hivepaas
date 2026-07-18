package entity

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
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

func (s *BasicAuth) GetResourceLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().GetResourceLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *BasicAuth) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentBasicAuthVersion {
		return false, nil
	}
	if setting.Version > CurrentBasicAuthVersion {
		return false, apperrors.Wrap(apperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentBasicAuthVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}

func (s *BasicAuth) Decrypt() error {
	_, err := s.Password.GetPlain()
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (s *Setting) AsBasicAuth() (*BasicAuth, error) {
	return parseSettingAs[*BasicAuth](s)
}

func (s *Setting) MustAsBasicAuth() *BasicAuth {
	return gofn.Must(s.AsBasicAuth())
}
