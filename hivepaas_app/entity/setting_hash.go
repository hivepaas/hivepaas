package entity

import (
	"strings"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/cryptoutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/randtoken"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/reflectutil"
)

const (
	hashingKeyLen    = 32
	hashingIteration = 1
)

type HashField struct {
	secret       string
	hashedSecret string
}

func (s *HashField) MarshalJSON() ([]byte, error) {
	hashedSecret, err := s.hash()
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	return reflectutil.UnsafeStrToBytes(gofn.StringWrap(hashedSecret, "\"")), nil
}

func (s *HashField) UnmarshalJSON(data []byte) error {
	s.hashedSecret = gofn.StringUnwrap(reflectutil.UnsafeBytesToStr(data), "\"")
	return nil
}

func (s *HashField) String() string {
	if s.secret != "" {
		return s.secret
	}
	return s.hashedSecret
}

func (s *HashField) IsHashed() bool {
	return s.hashedSecret != "" && s.secret == ""
}

func (s *HashField) Get() (string, error) {
	hashedSecret, err := s.hash()
	if err != nil {
		return "", apperrors.Wrap(err)
	}
	return hashedSecret, nil
}

func (s *HashField) MustGet() string {
	return gofn.Must(s.Get())
}

func (s *HashField) Set(value string) {
	if strings.HasPrefix(value, base.EncryptionSaltPrefix) {
		s.hashedSecret = value
		s.secret = ""
	} else {
		s.secret = value
		s.hashedSecret = ""
	}
}

func (s *HashField) VerifyHash(secret string) error {
	hashedSecret, salt := cryptoutil.UnpackSecret(s.hashedSecret)
	var matched bool
	if salt == "" {
		matched = hashedSecret == secret
	} else {
		matched = randtoken.VerifyHashHex(secret, hashedSecret, salt,
			hashingKeyLen, hashingIteration)
	}
	if !matched {
		return apperrors.Wrap(apperrors.ErrAPIKeyInvalid)
	}
	return nil
}

func (s *HashField) hash() (string, error) {
	if s.hashedSecret != "" {
		return s.hashedSecret, nil
	}
	hashedSecret, salt, err := randtoken.HashAsHex(s.secret, defaultSaltLen,
		hashingKeyLen, hashingIteration)
	if err != nil {
		return "", apperrors.Wrap(err)
	}
	s.hashedSecret = cryptoutil.PackSecret(hashedSecret, salt)
	return s.hashedSecret, nil
}

func NewHashField(value string) HashField {
	resp := HashField{}
	resp.Set(value)
	return resp
}
