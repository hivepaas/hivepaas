package entity

import (
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/base"
)

const (
	CurrentCloudProviderVersion = 1
)

var _ = registerSettingParser(base.SettingTypeCloudProvider, &cloudProviderParser{})

type cloudProviderParser struct {
}

func (s *cloudProviderParser) New() SettingData {
	return &CloudProvider{}
}

type CloudProvider struct {
	AWS *CloudProviderAWS `json:"aws,omitempty"`
}

type CloudProviderAWS struct {
	AccessKeyID string         `json:"accessKeyID"`
	SecretKey   EncryptedField `json:"secretKey"`
	Region      string         `json:"region,omitempty"`
}

func (s *CloudProvider) GetType() base.SettingType {
	return base.SettingTypeCloudProvider
}

func (s *CloudProvider) GetRefObjectIDs() *RefObjectIDs {
	return &RefObjectIDs{}
}

func (s *CloudProvider) MustDecrypt() *CloudProvider {
	if s.AWS != nil {
		s.AWS.SecretKey.MustGetPlain()
	}
	return s
}

func (s *CloudProvider) Migrate(setting *Setting) (hasChange bool, err error) {
	if CurrentCloudProviderVersion == setting.Version {
		return false, nil
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentCloudProviderVersion
	setting.MustSetData(s)
	return true, nil
}

func (s *Setting) AsCloudProvider() (*CloudProvider, error) {
	return parseSettingAs[*CloudProvider](s)
}

func (s *Setting) MustAsCloudProvider() *CloudProvider {
	return gofn.Must(s.AsCloudProvider())
}
