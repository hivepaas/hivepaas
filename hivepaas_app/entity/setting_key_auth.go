package entity

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

const (
	CurrentKeyAuthVersion = 1
)

var _ = registerSettingParser(base.SettingTypeKeyAuth, &keyAuthParser{})

type keyAuthParser struct {
}

func (s *keyAuthParser) New() SettingData {
	return &KeyAuth{}
}

// KeyAuth is a key id and its secret key: the credential of an S3 bucket, an
// object store, an API that signs its requests with an access key.
type KeyAuth struct {
	KeyID     string         `json:"keyId"`
	SecretKey EncryptedField `json:"secretKey"`
}

func (s *KeyAuth) GetType() base.SettingType {
	return base.SettingTypeKeyAuth
}

func (s *KeyAuth) GetRefObjectIDs() *RefObjectIDs {
	return &RefObjectIDs{}
}

func (s *KeyAuth) GetResourceLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().GetResourceLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *KeyAuth) Decrypt() error {
	_, err := s.SecretKey.GetPlain()
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (s *Setting) AsKeyAuth() (*KeyAuth, error) {
	return parseSettingAs[*KeyAuth](s)
}

func (s *Setting) MustAsKeyAuth() *KeyAuth {
	return gofn.Must(s.AsKeyAuth())
}
