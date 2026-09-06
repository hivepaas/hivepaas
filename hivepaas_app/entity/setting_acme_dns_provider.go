package entity

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
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

func (s *AcmeDnsProvider) GetResourceLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().GetResourceLinks(base.ResourceTypeSetting, setting.ID)
}

// SecretFieldFor returns the one encrypted secret the given provider kind stores,
// or nil when that kind holds no secret or is not the one populated here.
//
// Looking the field up by kind rather than by whichever sub-struct happens to be
// set is what makes it safe to pair a request against a stored setting: if the
// update switches provider, the stored side returns nil instead of a secret
// belonging to a different provider.
//
//nolint:gocognit
func (s *AcmeDnsProvider) SecretFieldFor(kind base.AcmeDnsProvider) *EncryptedField {
	switch kind {
	case base.AcmeDnsProviderAzure:
		if s.Azure != nil {
			return &s.Azure.ClientSecret
		}
	case base.AcmeDnsProviderBaiduCloud:
		if s.BaiduCloud != nil {
			return &s.BaiduCloud.SecretKey
		}
	case base.AcmeDnsProviderCloudflare:
		if s.Cloudflare != nil {
			return &s.Cloudflare.AuthToken
		}
	case base.AcmeDnsProviderDigitalOcean:
		if s.DigitalOcean != nil {
			return &s.DigitalOcean.AuthToken
		}
	case base.AcmeDnsProviderGCloud:
		if s.GCloud != nil {
			return &s.GCloud.ServiceAccount
		}
	case base.AcmeDnsProviderGoDaddy:
		if s.GoDaddy != nil {
			return &s.GoDaddy.APISecret
		}
	case base.AcmeDnsProviderHetzner:
		if s.Hetzner != nil {
			return &s.Hetzner.APIToken
		}
	case base.AcmeDnsProviderHuaweiCloud:
		if s.HuaweiCloud != nil {
			return &s.HuaweiCloud.SecretKey
		}
	case base.AcmeDnsProviderNamecheap:
		if s.Namecheap != nil {
			return &s.Namecheap.APIKey
		}
	case base.AcmeDnsProviderRFC2136:
		if s.RFC2136 != nil {
			return &s.RFC2136.TSIGSecret
		}
	case base.AcmeDnsProviderRoute53:
		if s.Route53 != nil {
			return &s.Route53.SecretAccessKey
		}
	case base.AcmeDnsProviderTencentCloud:
		if s.TencentCloud != nil {
			return &s.TencentCloud.SecretKey
		}
	case base.AcmeDnsProviderAcmeDNS:
	}
	return nil
}

// Decrypt reads every secret this provider holds, so a value that cannot be
// decrypted surfaces here rather than at the point of use.
//
// It walks the kinds rather than the sub-structs: SecretFieldFor already knows
// which field each kind stores, and restating that list here is how the two
// drift apart when a provider is added.
func (s *AcmeDnsProvider) Decrypt() error {
	for _, kind := range base.AllAcmeDnsProviders {
		field := s.SecretFieldFor(kind)
		if field == nil {
			continue
		}
		if _, err := field.GetPlain(); err != nil {
			return hperrors.Wrap(err)
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
