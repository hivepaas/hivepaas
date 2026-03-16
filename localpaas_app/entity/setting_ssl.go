package entity

import (
	"time"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
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
	Domain        string            `json:"domain"`
	Certificate   string            `json:"certificate"`
	PrivateKey    EncryptedField    `json:"privateKey"`
	KeySize       int               `json:"keySize"`
	Provider      base.SSLProvider  `json:"provider,omitempty"`
	Email         string            `json:"email"`
	AutoRenew     bool              `json:"autoRenew,omitempty"`
	RenewableFrom time.Time         `json:"renewableFrom,omitzero"`
	RenewableTo   time.Time         `json:"renewableTo,omitzero"`
	ExpireAt      time.Time         `json:"expireAt,omitzero"`
	NotifyWhen    timeutil.Duration `json:"notifyWhen,omitempty"`
}

func (s *SSL) GetType() base.SettingType {
	return base.SettingTypeSSL
}

func (s *SSL) GetRefObjectIDs() *RefObjectIDs {
	return &RefObjectIDs{}
}

func (s *SSL) MustDecrypt() *SSL {
	s.PrivateKey.MustGetPlain()
	return s
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
