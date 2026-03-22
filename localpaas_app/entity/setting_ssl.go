package entity

import (
	"time"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/base"
)

const (
	CurrentSSLVersion = 1
)

var _ = registerSettingParser(base.SettingTypeSSL, &sslParser{})

type sslParser struct {
}

func (s *sslParser) New() SettingData {
	return &SSL{}
}

type SSL struct {
	Domain        string                 `json:"domain"`
	Certificate   string                 `json:"certificate"`
	PrivateKey    EncryptedField         `json:"privateKey"`
	KeySize       int                    `json:"keySize"`
	Provider      base.SSLProvider       `json:"provider,omitempty"`
	Email         string                 `json:"email"`
	AutoRenew     bool                   `json:"autoRenew,omitempty"`
	RenewableFrom time.Time              `json:"renewableFrom,omitzero"`
	ExpireAt      time.Time              `json:"expireAt,omitzero"`
	NotifyFrom    time.Time              `json:"notifyFrom,omitzero"`
	Notification  *BaseEventNotification `json:"notification,omitempty"`
}

func (s *SSL) GetType() base.SettingType {
	return base.SettingTypeSSL
}

func (s *SSL) GetRefObjectIDs() *RefObjectIDs {
	refIDs := &RefObjectIDs{}
	if s.Notification != nil {
		refIDs.AddRefIDs(s.Notification.GetRefObjectIDs())
	}
	return refIDs
}

func (s *SSL) MustDecrypt() *SSL {
	s.PrivateKey.MustGetPlain()
	return s
}

func (s *SSL) IsRenewable() bool {
	return s.Provider == base.SSLProviderLetsEncrypt
}

func (s *SSL) Migrate(setting *Setting) (hasChange bool, err error) {
	if CurrentSSLVersion == setting.Version {
		return false, nil
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentSSLVersion
	setting.MustSetData(s)
	return true, nil
}

func (s *Setting) AsSSL() (*SSL, error) {
	return parseSettingAs[*SSL](s)
}

func (s *Setting) MustAsSSL() *SSL {
	return gofn.Must(s.AsSSL())
}
