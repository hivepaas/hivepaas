package entity

import (
	"strings"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/cryptoutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/reflectutil"
)

const (
	defaultSaltLen = 10
)

type EncryptedField struct {
	encrypted string
	decrypted string
}

func (s *EncryptedField) MarshalJSON() (res []byte, err error) {
	var encrypted string
	if config.Current.Secret != "" {
		encrypted, err = s.encrypt()
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
	} else {
		encrypted = s.encrypted
	}
	return reflectutil.UnsafeStrToBytes(gofn.StringWrap(encrypted, "\"")), nil
}

func (s *EncryptedField) UnmarshalJSON(data []byte) error {
	s.Set(gofn.StringUnwrap(reflectutil.UnsafeBytesToStr(data), "\""))
	return nil
}

func (s *EncryptedField) String() string {
	if s.decrypted != "" {
		return s.decrypted
	}
	return s.encrypted
}

func (s *EncryptedField) IsEncrypted() bool {
	return s.encrypted != "" && s.decrypted == ""
}

func (s *EncryptedField) GetPlain() (string, error) {
	decrypted, err := s.decrypt()
	if err != nil {
		return "", apperrors.Wrap(err)
	}
	return decrypted, nil
}

func (s *EncryptedField) GetEncrypted() (string, error) {
	encrypted, err := s.encrypt()
	if err != nil {
		return "", apperrors.Wrap(err)
	}
	return encrypted, nil
}

func (s *EncryptedField) Set(value string) {
	if strings.HasPrefix(value, base.EncryptionSaltPrefix) {
		s.encrypted = value
		s.decrypted = ""
	} else {
		s.decrypted = value
		s.encrypted = ""
	}
}

func (s *EncryptedField) Equal(enc *EncryptedField) (bool, error) {
	v1, err := s.GetPlain()
	if err != nil {
		return false, apperrors.Wrap(err)
	}
	v2, err := enc.GetPlain()
	if err != nil {
		return false, apperrors.Wrap(err)
	}
	return v1 == v2, nil
}

func (s *EncryptedField) encrypt() (string, error) {
	// TODO: should we use Mutex?
	if s.encrypted != "" {
		return s.encrypted, nil
	}
	if config.Current.Secret == "" {
		return "", apperrors.NewMissing("Encryption secret")
	}
	encrypted, err := cryptoutil.EncryptBase64(s.decrypted, defaultSaltLen, config.Current.Secret)
	if err != nil {
		return "", apperrors.Wrap(err)
	}
	s.encrypted = encrypted
	return encrypted, nil
}

func (s *EncryptedField) decrypt() (string, error) {
	// TODO: should we use Mutex?
	if s.decrypted != "" {
		return s.decrypted, nil
	}
	if config.Current.Secret == "" {
		return "", apperrors.NewMissing("Encryption secret")
	}
	decrypted, err := cryptoutil.DecryptBase64(s.encrypted, config.Current.Secret)
	if err != nil {
		return "", apperrors.Wrap(err)
	}
	s.decrypted = decrypted
	return decrypted, nil
}

func NewEncryptedField(value string) EncryptedField {
	resp := EncryptedField{}
	resp.Set(value)
	return resp
}
