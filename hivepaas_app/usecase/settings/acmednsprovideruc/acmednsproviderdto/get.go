package acmednsproviderdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/copier"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

const (
	maskedSecret = "****************"
)

type GetAcmeDnsProviderReq struct {
	settings.GetSettingReq
}

func NewGetAcmeDnsProviderReq() *GetAcmeDnsProviderReq {
	return &GetAcmeDnsProviderReq{}
}

func (req *GetAcmeDnsProviderReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.GetSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetAcmeDnsProviderResp struct {
	Meta *basedto.Meta        `json:"meta"`
	Data *AcmeDnsProviderResp `json:"data"`
}

type AcmeDnsProviderResp struct {
	*settings.BaseSettingResp
	Kind base.AcmeDnsProvider `json:"kind"`

	AcmeDNS      *AcmeDnsProviderAcmeDNSResp      `json:"acmeDns,omitempty"`
	Azure        *AcmeDnsProviderAzureResp        `json:"azure,omitempty"`
	BaiduCloud   *AcmeDnsProviderBaiduCloudResp   `json:"baiduCloud,omitempty"`
	Cloudflare   *AcmeDnsProviderCloudflareResp   `json:"cloudflare,omitempty"`
	DigitalOcean *AcmeDnsProviderDigitalOceanResp `json:"digitalOcean,omitempty"`
	GCloud       *AcmeDnsProviderGCloudResp       `json:"gCloud,omitempty"`
	GoDaddy      *AcmeDnsProviderGoDaddyResp      `json:"goDaddy,omitempty"`
	Hetzner      *AcmeDnsProviderHetznerResp      `json:"hetzner,omitempty"`
	HuaweiCloud  *AcmeDnsProviderHuaweiCloudResp  `json:"huaweiCloud,omitempty"`
	Namecheap    *AcmeDnsProviderNamecheapResp    `json:"namecheap,omitempty"`
	RFC2136      *AcmeDnsProviderRFC2136Resp      `json:"rfc2136,omitempty"`
	Route53      *AcmeDnsProviderRoute53Resp      `json:"route53,omitempty"`
	TencentCloud *AcmeDnsProviderTencentCloudResp `json:"tencentCloud,omitempty"`
	SecretMasked bool                             `json:"secretMasked,omitempty"`
}

type AcmeDnsProviderAcmeDNSResp struct {
	APIBase        string   `json:"apiBase"`
	AllowList      []string `json:"allowList"`
	StoragePath    string   `json:"storagePath"`
	StorageBaseURL string   `json:"storageBaseUrl"`
}

type AcmeDnsProviderAzureResp struct {
	ClientID          string `json:"clientId"`
	ClientSecret      string `json:"clientSecret"`
	SubscriptionID    string `json:"subscriptionId"`
	TenantID          string `json:"tenantId"`
	ResourceGroupName string `json:"resourceGroupName"`
}

func (resp *AcmeDnsProviderAzureResp) CopyClientSecret(field entity.EncryptedField) error {
	resp.ClientSecret = field.String()
	return nil
}

type AcmeDnsProviderBaiduCloudResp struct {
	AccessKey string `json:"accessKey"`
	SecretKey string `json:"secretKey"`
}

func (resp *AcmeDnsProviderBaiduCloudResp) CopySecretKey(field entity.EncryptedField) error {
	resp.SecretKey = field.String()
	return nil
}

type AcmeDnsProviderCloudflareResp struct {
	AuthToken string `json:"authToken"`
}

func (resp *AcmeDnsProviderCloudflareResp) CopyAuthToken(field entity.EncryptedField) error {
	resp.AuthToken = field.String()
	return nil
}

type AcmeDnsProviderDigitalOceanResp struct {
	AuthToken string `json:"authToken"`
}

func (resp *AcmeDnsProviderDigitalOceanResp) CopyAuthToken(field entity.EncryptedField) error {
	resp.AuthToken = field.String()
	return nil
}

type AcmeDnsProviderGCloudResp struct {
	ProjectID      string `json:"projectId"`
	ServiceAccount string `json:"serviceAccount"`
}

func (resp *AcmeDnsProviderGCloudResp) CopyServiceAccount(field entity.EncryptedField) error {
	resp.ServiceAccount = field.String()
	return nil
}

type AcmeDnsProviderGoDaddyResp struct {
	APIKey    string `json:"apiKey"`
	APISecret string `json:"apiSecret"`
}

func (resp *AcmeDnsProviderGoDaddyResp) CopyAPISecret(field entity.EncryptedField) error {
	resp.APISecret = field.String()
	return nil
}

type AcmeDnsProviderHetznerResp struct {
	APIToken string `json:"apiToken"`
}

func (resp *AcmeDnsProviderHetznerResp) CopyAPIToken(field entity.EncryptedField) error {
	resp.APIToken = field.String()
	return nil
}

type AcmeDnsProviderHuaweiCloudResp struct {
	AccessKey string `json:"accessKey"`
	SecretKey string `json:"secretKey"`
	Region    string `json:"region,omitempty"`
}

func (resp *AcmeDnsProviderHuaweiCloudResp) CopySecretKey(field entity.EncryptedField) error {
	resp.SecretKey = field.String()
	return nil
}

type AcmeDnsProviderNamecheapResp struct {
	APIUser string `json:"apiUser"`
	APIKey  string `json:"apiKey"`
}

func (resp *AcmeDnsProviderNamecheapResp) CopyAPIKey(field entity.EncryptedField) error {
	resp.APIKey = field.String()
	return nil
}

type AcmeDnsProviderRFC2136Resp struct {
	Nameserver    string `json:"nameserver"`
	TSIGKeyName   string `json:"tsigKeyName"`
	TSIGSecret    string `json:"tsigSecret"`
	TSIGAlgorithm string `json:"tsigAlgorithm"`
}

func (resp *AcmeDnsProviderRFC2136Resp) CopyTSIGSecret(field entity.EncryptedField) error {
	resp.TSIGSecret = field.String()
	return nil
}

type AcmeDnsProviderRoute53Resp struct {
	AccessKeyID     string `json:"accessKeyId"`
	SecretAccessKey string `json:"secretAccessKey"`
	HostedZoneID    string `json:"hostedZoneId,omitempty"`
	Region          string `json:"region,omitempty"`
}

func (resp *AcmeDnsProviderRoute53Resp) CopySecretAccessKey(field entity.EncryptedField) error {
	resp.SecretAccessKey = field.String()
	return nil
}

type AcmeDnsProviderTencentCloudResp struct {
	SecretID  string `json:"secretId"`
	SecretKey string `json:"secretKey"`
	Region    string `json:"region,omitempty"`
}

func (resp *AcmeDnsProviderTencentCloudResp) CopySecretKey(field entity.EncryptedField) error {
	resp.SecretKey = field.String()
	return nil
}

//nolint:gocognit,gocyclo
func TransformAcmeDnsProvider(
	setting *entity.Setting,
	_ *entity.RefObjects,
) (resp *AcmeDnsProviderResp, err error) {
	config := setting.MustAsAcmeDnsProvider()
	if err = copier.Copy(&resp, config); err != nil {
		return nil, apperrors.Wrap(err)
	}
	resp.Kind = base.AcmeDnsProvider(setting.Kind)

	resp.BaseSettingResp, err = settings.TransformSettingBase(setting)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	switch {
	case config.AcmeDNS != nil:
		resp.SecretMasked = false
	case config.Azure != nil:
		resp.SecretMasked = config.Azure.ClientSecret.IsEncrypted() || resp.Inherited
		if resp.SecretMasked {
			resp.Azure.ClientSecret = maskedSecret
		}
	case config.BaiduCloud != nil:
		resp.SecretMasked = config.BaiduCloud.SecretKey.IsEncrypted() || resp.Inherited
		if resp.SecretMasked {
			resp.BaiduCloud.SecretKey = maskedSecret
		}
	case config.Cloudflare != nil:
		resp.SecretMasked = config.Cloudflare.AuthToken.IsEncrypted() || resp.Inherited
		if resp.SecretMasked {
			resp.Cloudflare.AuthToken = maskedSecret
		}
	case config.DigitalOcean != nil:
		resp.SecretMasked = config.DigitalOcean.AuthToken.IsEncrypted() || resp.Inherited
		if resp.SecretMasked {
			resp.DigitalOcean.AuthToken = maskedSecret
		}
	case config.GCloud != nil:
		resp.SecretMasked = config.GCloud.ServiceAccount.IsEncrypted() || resp.Inherited
		if resp.SecretMasked {
			resp.GCloud.ServiceAccount = maskedSecret
		}
	case config.GoDaddy != nil:
		resp.SecretMasked = config.GoDaddy.APISecret.IsEncrypted() || resp.Inherited
		if resp.SecretMasked {
			resp.GoDaddy.APISecret = maskedSecret
		}
	case config.Hetzner != nil:
		resp.SecretMasked = config.Hetzner.APIToken.IsEncrypted() || resp.Inherited
		if resp.SecretMasked {
			resp.Hetzner.APIToken = maskedSecret
		}
	case config.HuaweiCloud != nil:
		resp.SecretMasked = config.HuaweiCloud.SecretKey.IsEncrypted() || resp.Inherited
		if resp.SecretMasked {
			resp.HuaweiCloud.SecretKey = maskedSecret
		}
	case config.Namecheap != nil:
		resp.SecretMasked = config.Namecheap.APIKey.IsEncrypted() || resp.Inherited
		if resp.SecretMasked {
			resp.Namecheap.APIKey = maskedSecret
		}
	case config.RFC2136 != nil:
		resp.SecretMasked = config.RFC2136.TSIGSecret.IsEncrypted() || resp.Inherited
		if resp.SecretMasked {
			resp.RFC2136.TSIGSecret = maskedSecret
		}
	case config.Route53 != nil:
		resp.SecretMasked = config.Route53.SecretAccessKey.IsEncrypted() || resp.Inherited
		if resp.SecretMasked {
			resp.Route53.SecretAccessKey = maskedSecret
		}
	case config.TencentCloud != nil:
		resp.SecretMasked = config.TencentCloud.SecretKey.IsEncrypted() || resp.Inherited
		if resp.SecretMasked {
			resp.TencentCloud.SecretKey = maskedSecret
		}
	}

	return resp, nil
}
