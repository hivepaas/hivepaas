package entity

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

const (
	CurrentDomainSettingsVersion = 1
)

var _ = registerSettingParser(base.SettingTypeDomainSettings, &domainSettingsParser{})

type domainSettingsParser struct {
}

func (s *domainSettingsParser) New() SettingData {
	return &DomainSettings{}
}

type DomainSettings struct {
	RootDomain     string              `json:"rootDomain"`
	AllowedDomains []string            `json:"allowedDomains"`
	CertSettings   *DomainCertSettings `json:"certSettings"`
}

type DomainCertSettings struct {
	CertType    base.SSLCertType  `json:"certType"`
	KeyType     base.SSLKeyType   `json:"keyType"`
	ValidPeriod timeutil.Duration `json:"validPeriod,omitempty"`
	Email       string            `json:"email"`
	AutoRenew   bool              `json:"autoRenew,omitempty"`
}

func (s *DomainSettings) GetType() base.SettingType {
	return base.SettingTypeDomainSettings
}

func (s *DomainSettings) GetRefObjectIDs() *RefObjectIDs {
	refIDs := &RefObjectIDs{}
	return refIDs
}

func (s *DomainSettings) GetResourceLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().GetResourceLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *Setting) AsDomainSettings() (*DomainSettings, error) {
	return parseSettingAs[*DomainSettings](s)
}

func (s *Setting) MustAsDomainSettings() *DomainSettings {
	return gofn.Must(s.AsDomainSettings())
}
