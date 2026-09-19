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

	// AutoObtain asks HivePaaS to get a certificate for a domain nothing already
	// covers, rather than leaving the app on plain HTTP until somebody creates one
	// by hand. It is off where nothing can be obtained - an installation not
	// reachable from the internet, or one whose domains are internal names - and
	// an attempt there would only spend a certificate authority's patience.
	AutoObtain bool `json:"autoObtain,omitempty"`
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
