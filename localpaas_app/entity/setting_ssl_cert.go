package entity

import (
	"time"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
)

const (
	CurrentSSLCertVersion = 1
)

var _ = registerSettingParser(base.SettingTypeSSLCert, &sslCertParser{})

type sslCertParser struct {
}

func (s *sslCertParser) New() SettingData {
	return &SSLCert{}
}

type SSLCert struct {
	CertType      base.SSLCertType       `json:"certType"`
	Domain        string                 `json:"domain"`
	Certificate   string                 `json:"certificate"`
	PrivateKey    EncryptedField         `json:"privateKey"`
	KeyType       base.SSLKeyType        `json:"keyType"`
	Email         string                 `json:"email"`
	BaseFilename  string                 `json:"baseFilename,omitempty"`
	AutoRenew     bool                   `json:"autoRenew,omitempty"`
	RenewableFrom time.Time              `json:"renewableFrom,omitzero"`
	ExpireAt      time.Time              `json:"expireAt,omitzero"`
	NotifyFrom    time.Time              `json:"notifyFrom,omitzero"`
	Notification  *BaseEventNotification `json:"notification,omitempty"`
}

func (s *SSLCert) GetType() base.SettingType {
	return base.SettingTypeSSLCert
}

func (s *SSLCert) GetRefObjectIDs() *RefObjectIDs {
	refIDs := &RefObjectIDs{}
	if s.Notification != nil {
		refIDs.AddRefIDs(s.Notification.GetRefObjectIDs())
	}
	return refIDs
}

func (s *SSLCert) MustDecrypt() *SSLCert {
	s.PrivateKey.MustGetPlain()
	return s
}

func (s *SSLCert) IsRenewable() bool {
	return s.CertType == base.SSLCertTypeLetsEncrypt || s.CertType == base.SSLCertTypeSelfSigned
}

func (s *SSLCert) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentSSLCertVersion {
		return false, nil
	}
	if setting.Version > CurrentSSLCertVersion {
		return false, apperrors.New(apperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentSSLCertVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}

func (s *Setting) AsSSLCert() (*SSLCert, error) {
	return parseSettingAs[*SSLCert](s)
}

func (s *Setting) MustAsSSLCert() *SSLCert {
	return gofn.Must(s.AsSSLCert())
}
