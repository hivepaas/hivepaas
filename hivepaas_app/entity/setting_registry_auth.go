package entity

import (
	"github.com/moby/moby/api/types/registry"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
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

// RegistryAuthManagedBySystemRegistry marks, in RegistryAuth.ManagedBy, the
// credential the system registry created for itself. It is how the registry
// knows its own credential again whatever it is called or addressed at now: an
// operator may rename it, and the registry's domain may change between one
// switch-on and the next. It is in the data rather than in Setting.Kind, which
// a registry auth keeps its address in.
const RegistryAuthManagedBySystemRegistry = "system-registry"

type RegistryAuth struct {
	Username string         `json:"username"`
	Password EncryptedField `json:"password"`
	Address  string         `json:"address"`
	Readonly bool           `json:"readonly,omitempty"`
	// ManagedBy names what created the credential and keeps it, when that is not
	// an operator. It cannot be set or cleared through the API: an edit carries
	// it over.
	ManagedBy string `json:"managedBy,omitempty"`
}

func (s *RegistryAuth) GetType() base.SettingType {
	return base.SettingTypeRegistryAuth
}

func (s *RegistryAuth) GetRefObjectIDs() *RefObjectIDs {
	return &RefObjectIDs{}
}

func (s *RegistryAuth) GetResourceLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().GetResourceLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *RegistryAuth) Decrypt() error {
	_, err := s.Password.GetPlain()
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (s *RegistryAuth) GenerateAuthHeader() (string, error) {
	password, err := s.Password.GetPlain()
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	h, err := docker.GenerateAuthHeader(&registry.AuthConfig{
		Username:      s.Username,
		Password:      password,
		ServerAddress: s.Address,
	})
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	return h, nil
}

func (s *Setting) AsRegistryAuth() (*RegistryAuth, error) {
	return parseSettingAs[*RegistryAuth](s)
}

func (s *Setting) MustAsRegistryAuth() *RegistryAuth {
	return gofn.Must(s.AsRegistryAuth())
}
