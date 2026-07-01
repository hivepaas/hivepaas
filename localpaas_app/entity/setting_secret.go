package entity

import (
	"encoding/base64"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/pkg/fileutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/reflectutil"
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
	Key      string          `json:"key"`
	Value    EncryptedField  `json:"value"`
	Base64   bool            `json:"base64,omitempty"`
	SwarmRef *SwarmSecretRef `json:"swarmRef,omitempty"`
}

type SwarmSecretRef struct {
	File       *SwarmRefFileTarget `json:"file"`
	SecretID   string              `json:"secretId"`
	SecretName string              `json:"secretName"`
}

type SwarmRefFileTarget struct {
	Name string            `json:"name,omitempty"`
	UID  string            `json:"uid,omitempty"`
	GID  string            `json:"gid,omitempty"`
	Mode fileutil.FileMode `json:"mode,omitempty"`
}

func (s *Secret) GetType() base.SettingType {
	return base.SettingTypeSecret
}

func (s *Secret) GetRefObjectIDs() *RefObjectIDs {
	return &RefObjectIDs{}
}

func (s *Secret) CalcResLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().CalcResLinks(base.ResourceTypeSetting, setting.ID)
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

func (s *Secret) Decrypt() error {
	_, err := s.Value.GetPlain()
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}

func (s *Secret) ValueAsBytes() ([]byte, error) {
	plain, err := s.Value.GetPlain()
	if err != nil {
		return nil, apperrors.New(err)
	}
	if s.Base64 {
		plainBytes, err := base64.StdEncoding.DecodeString(plain)
		if err != nil {
			return nil, apperrors.New(err)
		}
		return plainBytes, nil
	}
	return reflectutil.UnsafeStrToBytes(plain), nil
}

func (s *Secret) ValueSize() (int64, error) {
	plain, err := s.Value.GetPlain()
	if err != nil {
		return 0, apperrors.New(err)
	}
	return int64(len(plain)), nil //nolint:gosec
}

func (s *Setting) AsSecret() (*Secret, error) {
	return parseSettingAs[*Secret](s)
}

func (s *Setting) MustAsSecret() *Secret {
	return gofn.Must(s.AsSecret())
}
