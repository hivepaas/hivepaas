package entity

import (
	"github.com/moby/moby/api/types/registry"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/services/docker"
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

func (s *RegistryAuth) CalcResLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().CalcResLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *RegistryAuth) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentRegistryAuthVersion {
		return false, nil
	}
	if setting.Version > CurrentRegistryAuthVersion {
		return false, apperrors.New(apperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentRegistryAuthVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}

func (s *RegistryAuth) Decrypt() error {
	_, err := s.Password.GetPlain()
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}

func (s *RegistryAuth) GenerateAuthHeader() (string, error) {
	password, err := s.Password.GetPlain()
	if err != nil {
		return "", apperrors.New(err)
	}
	h, err := docker.GenerateAuthHeader(&registry.AuthConfig{
		Username:      s.Username,
		Password:      password,
		ServerAddress: s.Address,
	})
	if err != nil {
		return "", apperrors.New(err)
	}
	return h, nil
}

func (s *Setting) AsRegistryAuth() (*RegistryAuth, error) {
	return parseSettingAs[*RegistryAuth](s)
}

func (s *Setting) MustAsRegistryAuth() *RegistryAuth {
	return gofn.Must(s.AsRegistryAuth())
}
