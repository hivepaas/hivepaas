package entity

import (
	"encoding/base64"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/reflectutil"
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

// Secret is a value an app reads through ${secrets.NAME}, or mounts as a file
// through a setting mount (app-setting-mount).
type Secret struct {
	Key    string         `json:"key"`
	Value  EncryptedField `json:"value"`
	Base64 bool           `json:"base64,omitempty"`
}

func (s *Secret) GetType() base.SettingType {
	return base.SettingTypeSecret
}

func (s *Secret) GetRefObjectIDs() *RefObjectIDs {
	return &RefObjectIDs{}
}

func (s *Secret) GetResourceLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().GetResourceLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *Secret) Decrypt() error {
	_, err := s.Value.GetPlain()
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (s *Secret) ValueAsBytes() ([]byte, error) {
	plain, err := s.Value.GetPlain()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if s.Base64 {
		plainBytes, err := base64.StdEncoding.DecodeString(plain)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		return plainBytes, nil
	}
	return reflectutil.UnsafeStrToBytes(plain), nil
}

func (s *Secret) ValueSize() (int64, error) {
	plain, err := s.Value.GetPlain()
	if err != nil {
		return 0, hperrors.Wrap(err)
	}
	return int64(len(plain)), nil //nolint:gosec
}

func (s *Setting) AsSecret() (*Secret, error) {
	return parseSettingAs[*Secret](s)
}

func (s *Setting) MustAsSecret() *Secret {
	return gofn.Must(s.AsSecret())
}
