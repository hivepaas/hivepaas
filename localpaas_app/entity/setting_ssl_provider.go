package entity

import (
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
)

const (
	CurrentSSLProviderVersion = 1
)

var _ = registerSettingParser(base.SettingTypeSSLProvider, &sslProviderParser{})

type sslProviderParser struct {
}

func (s *sslProviderParser) New() SettingData {
	return &SSLProvider{}
}

type SSLProvider struct {
	LetsEncrypt       *SSLProviderLetsEncrypt `json:"letsEncrypt,omitempty"`
	ZeroSSL           *SSLProviderZeroSSL     `json:"zeroSSL,omitempty"`
	GoogleTrust       *SSLProviderGoogleTrust `json:"googleTrust,omitempty"`
	Email             string                  `json:"email"`
	DefaultKeyType    base.SSLKeyType         `json:"defaultKeyType,omitempty"`
	SupportedKeyTypes []base.SSLKeyType       `json:"supportedKeyTypes,omitempty"`
}

type SSLProviderLetsEncrypt struct {
}

type SSLProviderZeroSSL struct {
	EABKid     string         `json:"eabKid"`
	EABHmacKey EncryptedField `json:"eabHmacKey"`
}

type SSLProviderGoogleTrust struct {
	EABKid     string         `json:"eabKid"`
	EABHmacKey EncryptedField `json:"eabHmacKey"`
}

func (s *SSLProvider) GetType() base.SettingType {
	return base.SettingTypeSSLProvider
}

func (s *SSLProvider) GetRefObjectIDs() *RefObjectIDs {
	refIDs := &RefObjectIDs{}
	return refIDs
}

func (s *SSLProvider) MustDecrypt() *SSLProvider {
	if s.ZeroSSL != nil {
		s.ZeroSSL.EABHmacKey.MustGetPlain()
	}
	if s.GoogleTrust != nil {
		s.GoogleTrust.EABHmacKey.MustGetPlain()
	}
	return s
}

func (s *SSLProvider) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentSSLProviderVersion {
		return false, nil
	}
	if setting.Version > CurrentSSLProviderVersion {
		return false, apperrors.New(apperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentSSLProviderVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}

func (s *Setting) AsSSLProvider() (*SSLProvider, error) {
	return parseSettingAs[*SSLProvider](s)
}

func (s *Setting) MustAsSSLProvider() *SSLProvider {
	return gofn.Must(s.AsSSLProvider())
}
