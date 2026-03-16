package entity

import (
	"github.com/docker/docker/api/types/registry"
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
)

const (
	CurrentRegistryAuthVersion = 1
)

var _ = registerSettingParser(base.SettingTypeRegistryAuth, &registryAuthParser{})

type registryAuthParser struct {
}

func (s *registryAuthParser) New() SettingData {
	return &RegistryAuth{}
}

type RegistryAuth struct {
	Username string         `json:"username"`
	Password EncryptedField `json:"password"`
	Address  string         `json:"address"`
	Readonly bool           `json:"readonly,omitempty"`
}

func (s *RegistryAuth) GetType() base.SettingType {
	return base.SettingTypeRegistryAuth
}

func (s *RegistryAuth) GetRefObjectIDs() *RefObjectIDs {
	return &RefObjectIDs{}
}

func (s *RegistryAuth) MustDecrypt() *RegistryAuth {
	s.Password.MustGetPlain()
	return s
}

func (s *RegistryAuth) GenerateAuthHeader() (string, error) {
	password, err := s.Password.GetPlain()
	if err != nil {
		return "", apperrors.Wrap(err)
	}
	h, err := registry.EncodeAuthConfig(registry.AuthConfig{
		Username:      s.Username,
		Password:      password,
		ServerAddress: s.Address,
	})
	if err != nil {
		return "", apperrors.Wrap(err)
	}
	return h, nil
}

func (s *RegistryAuth) Migrate(setting *Setting) (hasChange bool, err error) {
	if CurrentRegistryAuthVersion == setting.Version {
		return false, nil
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentRegistryAuthVersion
	setting.MustSetData(s)
	return true, nil
}

func (s *Setting) AsRegistryAuth() (*RegistryAuth, error) {
	return parseSettingAs[*RegistryAuth](s)
}

func (s *Setting) MustAsRegistryAuth() *RegistryAuth {
	return gofn.Must(s.AsRegistryAuth())
}
