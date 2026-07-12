package entity

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

const (
	CurrentAcmeDnsProviderVersion = 1
)

var _ = registerSettingParser(base.SettingTypeAcmeDnsProvider, &acmeDnsProviderParser{})

type acmeDnsProviderParser struct {
}

func (s *acmeDnsProviderParser) New() SettingData {
	return &AcmeDnsProvider{}
}

type AcmeDnsProvider struct {
	AcmeDNS      *AcmeDnsProviderAcmeDNS      `json:"acmeDns,omitempty"`
	Azure        *AcmeDnsProviderAzure        `json:"azure,omitempty"`
	BaiduCloud   *AcmeDnsProviderBaiduCloud   `json:"baiduCloud,omitempty"`
	Cloudflare   *AcmeDnsProviderCloudflare   `json:"cloudflare,omitempty"`
	DigitalOcean *AcmeDnsProviderDigitalOcean `json:"digitalOcean,omitempty"`
	GCloud       *AcmeDnsProviderGCloud       `json:"gCloud,omitempty"`
	GoDaddy      *AcmeDnsProviderGoDaddy      `json:"goDaddy,omitempty"`
	Hetzner      *AcmeDnsProviderHetzner      `json:"hetzner,omitempty"`
	HuaweiCloud  *AcmeDnsProviderHuaweiCloud  `json:"huaweiCloud,omitempty"`
	Namecheap    *AcmeDnsProviderNamecheap    `json:"namecheap,omitempty"`
	RFC2136      *AcmeDnsProviderRFC2136      `json:"rfc2136,omitempty"`
	Route53      *AcmeDnsProviderRoute53      `json:"route53,omitempty"`
	TencentCloud *AcmeDnsProviderTencentCloud `json:"tencentCloud,omitempty"`
}

type AcmeDnsProviderAcmeDNS struct {
	APIBase        string   `json:"apiBase"`
	AllowList      []string `json:"allowList"`
	StoragePath    string   `json:"storagePath"`
	StorageBaseURL string   `json:"storageBaseUrl"`
}

type AcmeDnsProviderAzure struct {
	ClientID          string         `json:"clientId"`
	ClientSecret      EncryptedField `json:"clientSecret"`
	SubscriptionID    string         `json:"subscriptionId"`
	TenantID          string         `json:"tenantId"`
	ResourceGroupName string         `json:"resourceGroupName"`
}

type AcmeDnsProviderBaiduCloud struct {
	AccessKey string         `json:"accessKey"`
	SecretKey EncryptedField `json:"secretKey"`
}

type AcmeDnsProviderCloudflare struct {
	AuthToken EncryptedField `json:"authToken"`
}

type AcmeDnsProviderDigitalOcean struct {
	AuthToken EncryptedField `json:"authToken"`
}

type AcmeDnsProviderGCloud struct {
	ProjectID      string         `json:"projectId"`
	ServiceAccount EncryptedField `json:"serviceAccount"`
}

type AcmeDnsProviderGoDaddy struct {
	APIKey    string         `json:"apiKey"`
	APISecret EncryptedField `json:"apiSecret"`
}

type AcmeDnsProviderHetzner struct {
	APIToken EncryptedField `json:"apiToken"`
}

type AcmeDnsProviderHuaweiCloud struct {
	AccessKey string         `json:"accessKey"`
	SecretKey EncryptedField `json:"secretKey"`
	Region    string         `json:"region,omitempty"`
}

type AcmeDnsProviderNamecheap struct {
	APIUser string         `json:"apiUser"`
	APIKey  EncryptedField `json:"apiKey"`
}

type AcmeDnsProviderRFC2136 struct {
	Nameserver    string         `json:"nameserver"`
	TSIGKeyName   string         `json:"tsigKeyName"`
	TSIGSecret    EncryptedField `json:"tsigSecret"`
	TSIGAlgorithm string         `json:"tsigAlgorithm"`
}

type AcmeDnsProviderRoute53 struct {
	AccessKeyID     string         `json:"accessKeyId"`
	SecretAccessKey EncryptedField `json:"secretAccessKey"`
	HostedZoneID    string         `json:"hostedZoneId,omitempty"`
	Region          string         `json:"region,omitempty"`
}

type AcmeDnsProviderTencentCloud struct {
	SecretID  string         `json:"secretId"`
	SecretKey EncryptedField `json:"secretKey"`
	Region    string         `json:"region,omitempty"`
}

func (s *AcmeDnsProvider) GetType() base.SettingType {
	return base.SettingTypeAcmeDnsProvider
}

func (s *AcmeDnsProvider) GetRefObjectIDs() *RefObjectIDs {
	refIDs := &RefObjectIDs{}
	return refIDs
}

func (s *AcmeDnsProvider) CalcResLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().CalcResLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *AcmeDnsProvider) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentAcmeDnsProviderVersion {
		return false, nil
	}
	if setting.Version > CurrentAcmeDnsProviderVersion {
		return false, apperrors.Wrap(apperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentAcmeDnsProviderVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}

//nolint:gocognit
func (s *AcmeDnsProvider) Decrypt() error {
	if s.Azure != nil {
		_, err := s.Azure.ClientSecret.GetPlain()
		if err != nil {
			return apperrors.Wrap(err)
		}
	}
	if s.BaiduCloud != nil {
		_, err := s.BaiduCloud.SecretKey.GetPlain()
		if err != nil {
			return apperrors.Wrap(err)
		}
	}
	if s.Cloudflare != nil {
		_, err := s.Cloudflare.AuthToken.GetPlain()
		if err != nil {
			return apperrors.Wrap(err)
		}
	}
	if s.DigitalOcean != nil {
		_, err := s.DigitalOcean.AuthToken.GetPlain()
		if err != nil {
			return apperrors.Wrap(err)
		}
	}
	if s.GCloud != nil {
		_, err := s.GCloud.ServiceAccount.GetPlain()
		if err != nil {
			return apperrors.Wrap(err)
		}
	}
	if s.GoDaddy != nil {
		_, err := s.GoDaddy.APISecret.GetPlain()
		if err != nil {
			return apperrors.Wrap(err)
		}
	}
	if s.Hetzner != nil {
		_, err := s.Hetzner.APIToken.GetPlain()
		if err != nil {
			return apperrors.Wrap(err)
		}
	}
	if s.HuaweiCloud != nil {
		_, err := s.HuaweiCloud.SecretKey.GetPlain()
		if err != nil {
			return apperrors.Wrap(err)
		}
	}
	if s.Namecheap != nil {
		_, err := s.Namecheap.APIKey.GetPlain()
		if err != nil {
			return apperrors.Wrap(err)
		}
	}
	if s.RFC2136 != nil {
		_, err := s.RFC2136.TSIGSecret.GetPlain()
		if err != nil {
			return apperrors.Wrap(err)
		}
	}
	if s.Route53 != nil {
		_, err := s.Route53.SecretAccessKey.GetPlain()
		if err != nil {
			return apperrors.Wrap(err)
		}
	}
	if s.TencentCloud != nil {
		_, err := s.TencentCloud.SecretKey.GetPlain()
		if err != nil {
			return apperrors.Wrap(err)
		}
	}
	return nil
}

func (s *Setting) AsAcmeDnsProvider() (*AcmeDnsProvider, error) {
	return parseSettingAs[*AcmeDnsProvider](s)
}

func (s *Setting) MustAsAcmeDnsProvider() *AcmeDnsProvider {
	return gofn.Must(s.AsAcmeDnsProvider())
}
