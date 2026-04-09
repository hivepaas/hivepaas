package entity

import (
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
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
	RootDomain   string              `json:"rootDomain"`
	CertSettings *DomainCertSettings `json:"certSettings"`
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

func (s *DomainSettings) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentDomainSettingsVersion {
		return false, nil
	}
	if setting.Version > CurrentDomainSettingsVersion {
		return false, apperrors.New(apperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentDomainSettingsVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}

func (s *Setting) AsDomainSettings() (*DomainSettings, error) {
	return parseSettingAs[*DomainSettings](s)
}

func (s *Setting) MustAsDomainSettings() *DomainSettings {
	return gofn.Must(s.AsDomainSettings())
}
