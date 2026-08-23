package entity

import (
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
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
	CertType      base.SSLCertType `json:"certType"`
	Provider      ObjectID         `json:"provider,omitzero"`
	Domain        string           `json:"domain"`
	Certificate   string           `json:"certificate"`
	PrivateKey    EncryptedField   `json:"privateKey"`
	CACertificate string           `json:"caCertificate,omitempty"`
	KeyType       base.SSLKeyType  `json:"keyType"`

	ValidPeriod   timeutil.Duration      `json:"validPeriod"`
	Email         string                 `json:"email"`
	BaseFilename  string                 `json:"baseFilename,omitempty"`
	AutoRenew     bool                   `json:"autoRenew,omitempty"`
	AcmeProvider  ObjectID               `json:"acmeProvider,omitzero"`
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
	if s.Provider.ID != "" {
		refIDs.RefSettingIDs = append(refIDs.RefSettingIDs, s.Provider.ID)
	}
	if s.AcmeProvider.ID != "" {
		refIDs.RefSettingIDs = append(refIDs.RefSettingIDs, s.AcmeProvider.ID)
	}
	if s.Notification != nil {
		refIDs.AddRefIDs(s.Notification.GetRefObjectIDs())
	}
	return refIDs
}

func (s *SSLCert) GetResourceLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().GetResourceLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *SSLCert) Decrypt() error {
	_, err := s.PrivateKey.GetPlain()
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (s *SSLCert) IsRenewable() bool {
	return s.CertType != base.SSLCertTypeCustom
}

func (s *Setting) AsSSLCert() (*SSLCert, error) {
	return parseSettingAs[*SSLCert](s)
}

func (s *Setting) MustAsSSLCert() *SSLCert {
	return gofn.Must(s.AsSSLCert())
}
