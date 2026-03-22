package entity

import (
	"time"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
)

const (
	CurrentSSLRenewalVersion = 1
)

var _ = registerSettingParser(base.SettingTypeSSLRenewal, &sslRenewalParser{})

type sslRenewalParser struct {
}

func (s *sslRenewalParser) New() SettingData {
	return &SSLRenewal{}
}

type SSLRenewal struct {
	ScheduleInterval timeutil.Duration `json:"scheduleInterval"`
	ScheduleFrom     time.Time         `json:"scheduleFrom"`
}

func (s *SSLRenewal) GetType() base.SettingType {
	return base.SettingTypeSSLRenewal
}

func (s *SSLRenewal) GetRefObjectIDs() *RefObjectIDs {
	return &RefObjectIDs{}
}

func (s *SSLRenewal) Migrate(setting *Setting) (hasChange bool, err error) {
	if CurrentSSLRenewalVersion == setting.Version {
		return false, nil
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentSSLRenewalVersion
	setting.MustSetData(s)
	return true, nil
}

func (s *Setting) AsSSLRenewal() (*SSLRenewal, error) {
	return parseSettingAs[*SSLRenewal](s)
}

func (s *Setting) MustAsSSLRenewal() *SSLRenewal {
	return gofn.Must(s.AsSSLRenewal())
}
