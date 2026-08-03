package entity

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
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
	Schedule     SchedJobSchedule       `json:"schedule"`
	Notification *BaseEventNotification `json:"notification,omitempty"`
}

func (s *SSLRenewal) GetType() base.SettingType {
	return base.SettingTypeSSLRenewal
}

func (s *SSLRenewal) GetRefObjectIDs() *RefObjectIDs {
	refIDs := &RefObjectIDs{}
	if s.Notification != nil {
		refIDs.AddRefIDs(s.Notification.GetRefObjectIDs())
	}
	return refIDs
}

func (s *SSLRenewal) GetResourceLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().GetResourceLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *Setting) AsSSLRenewal() (*SSLRenewal, error) {
	return parseSettingAs[*SSLRenewal](s)
}

func (s *Setting) MustAsSSLRenewal() *SSLRenewal {
	return gofn.Must(s.AsSSLRenewal())
}
