package entity

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
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

func (s *SSLProvider) GetResourceLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().GetResourceLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *SSLProvider) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentSSLProviderVersion {
		return false, nil
	}
	if setting.Version > CurrentSSLProviderVersion {
		return false, apperrors.Wrap(apperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentSSLProviderVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}

func (s *SSLProvider) Decrypt() error {
	if s.ZeroSSL != nil {
		_, err := s.ZeroSSL.EABHmacKey.GetPlain()
		if err != nil {
			return apperrors.Wrap(err)
		}
	}
	if s.GoogleTrust != nil {
		_, err := s.GoogleTrust.EABHmacKey.GetPlain()
		if err != nil {
			return apperrors.Wrap(err)
		}
	}
	return nil
}
func (s *Setting) AsSSLProvider() (*SSLProvider, error) {
	return parseSettingAs[*SSLProvider](s)
}

func (s *Setting) MustAsSSLProvider() *SSLProvider {
	return gofn.Must(s.AsSSLProvider())
}
