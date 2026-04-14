package entity

import (
	"encoding/base64"
	"os"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
)

const (
	CurrentSecretVersion = 1
)

var _ = registerSettingParser(base.SettingTypeSecret, &secretParser{})

type secretParser struct {
}

func (s *secretParser) New() SettingData {
	return &Secret{}
}

type Secret struct {
	Key      string          `json:"k"`
	Value    EncryptedField  `json:"v"`
	Base64   bool            `json:"b64"`
	SwarmRef *SwarmSecretRef `json:"swarmRef,omitempty"`
}

type SwarmSecretRef struct {
	File       *SwarmSecretRefFileTarget `json:"file"`
	SecretID   string                    `json:"secretId"`
	SecretName string                    `json:"secretName"`
}

type SwarmSecretRefFileTarget struct {
	Name string      `json:"name,omitempty"`
	UID  string      `json:"uid,omitempty"`
	GID  string      `json:"gid,omitempty"`
	Mode os.FileMode `json:"mode,omitempty"`
}

func (s *Secret) GetType() base.SettingType {
	return base.SettingTypeSecret
}

func (s *Secret) GetRefObjectIDs() *RefObjectIDs {
	return &RefObjectIDs{}
}

func (s *Secret) MustDecrypt() *Secret {
	s.Value.MustGetPlain()
	return s
}

func (s *Secret) ValueAsBytes() []byte {
	plain := s.Value.MustGetPlain()
	if s.Base64 {
		return gofn.Must(base64.StdEncoding.DecodeString(plain))
	}
	return []byte(plain)
}

func (s *Secret) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentSecretVersion {
		return false, nil
	}
	if setting.Version > CurrentSecretVersion {
		return false, apperrors.New(apperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentSecretVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}

func (s *Setting) AsSecret() (*Secret, error) {
	return parseSettingAs[*Secret](s)
}

func (s *Setting) MustAsSecret() *Secret {
	return gofn.Must(s.AsSecret())
}
