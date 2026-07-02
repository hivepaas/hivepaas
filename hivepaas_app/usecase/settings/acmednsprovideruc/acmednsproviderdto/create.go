package acmednsproviderdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

const (
	len10000 = 10000
	len1000  = 1000
	len200   = 200
)

type CreateAcmeDnsProviderReq struct {
	settings.CreateSettingReq
	*AcmeDnsProviderBaseReq
}

type AcmeDnsProviderBaseReq struct {
	Name string               `json:"name"`
	Kind base.AcmeDnsProvider `json:"kind"`

	AcmeDNS      *AcmeDnsProviderAcmeDNSReq      `json:"acmeDns"`
	Azure        *AcmeDnsProviderAzureReq        `json:"azure"`
	BaiduCloud   *AcmeDnsProviderBaiduCloudReq   `json:"baiduCloud"`
	Cloudflare   *AcmeDnsProviderCloudflareReq   `json:"cloudflare"`
	DigitalOcean *AcmeDnsProviderDigitalOceanReq `json:"digitalOcean"`
	GCloud       *AcmeDnsProviderGCloudReq       `json:"gCloud"`
	GoDaddy      *AcmeDnsProviderGoDaddyReq      `json:"goDaddy"`
	Hetzner      *AcmeDnsProviderHetznerReq      `json:"hetzner"`
	HuaweiCloud  *AcmeDnsProviderHuaweiCloudReq  `json:"huaweiCloud"`
	Namecheap    *AcmeDnsProviderNamecheapReq    `json:"namecheap"`
	RFC2136      *AcmeDnsProviderRFC2136Req      `json:"rfc2136"`
	Route53      *AcmeDnsProviderRoute53Req      `json:"route53"`
	TencentCloud *AcmeDnsProviderTencentCloudReq `json:"tencentCloud"`
}

func (req *AcmeDnsProviderBaseReq) ToEntity() *entity.AcmeDnsProvider {
	acmeDnsProvider := &entity.AcmeDnsProvider{}
	switch req.Kind {
	case base.AcmeDnsProviderAcmeDNS:
		if req.AcmeDNS != nil {
			acmeDnsProvider.AcmeDNS = &entity.AcmeDnsProviderAcmeDNS{
				APIBase:        req.AcmeDNS.APIBase,
				AllowList:      req.AcmeDNS.AllowList,
				StoragePath:    req.AcmeDNS.StoragePath,
				StorageBaseURL: req.AcmeDNS.StorageBaseURL,
			}
		}
	case base.AcmeDnsProviderAzure:
		if req.Azure != nil {
			acmeDnsProvider.Azure = &entity.AcmeDnsProviderAzure{
				ClientID:          req.Azure.ClientID,
				ClientSecret:      entity.NewEncryptedField(req.Azure.ClientSecret),
				SubscriptionID:    req.Azure.SubscriptionID,
				TenantID:          req.Azure.TenantID,
				ResourceGroupName: req.Azure.ResourceGroupName,
			}
		}
	case base.AcmeDnsProviderBaiduCloud:
		if req.BaiduCloud != nil {
			acmeDnsProvider.BaiduCloud = &entity.AcmeDnsProviderBaiduCloud{
				AccessKey: req.BaiduCloud.AccessKey,
				SecretKey: entity.NewEncryptedField(req.BaiduCloud.SecretKey),
			}
		}
	case base.AcmeDnsProviderCloudflare:
		if req.Cloudflare != nil {
			acmeDnsProvider.Cloudflare = &entity.AcmeDnsProviderCloudflare{
				AuthToken: entity.NewEncryptedField(req.Cloudflare.AuthToken),
			}
		}
	case base.AcmeDnsProviderDigitalOcean:
		if req.DigitalOcean != nil {
			acmeDnsProvider.DigitalOcean = &entity.AcmeDnsProviderDigitalOcean{
				AuthToken: entity.NewEncryptedField(req.DigitalOcean.AuthToken),
			}
		}
	case base.AcmeDnsProviderGCloud:
		if req.GCloud != nil {
			acmeDnsProvider.GCloud = &entity.AcmeDnsProviderGCloud{
				ProjectID:      req.GCloud.ProjectID,
				ServiceAccount: entity.NewEncryptedField(req.GCloud.ServiceAccount),
			}
		}
	case base.AcmeDnsProviderGoDaddy:
		if req.GoDaddy != nil {
			acmeDnsProvider.GoDaddy = &entity.AcmeDnsProviderGoDaddy{
				APIKey:    req.GoDaddy.APIKey,
				APISecret: entity.NewEncryptedField(req.GoDaddy.APISecret),
			}
		}
	case base.AcmeDnsProviderHetzner:
		if req.Hetzner != nil {
			acmeDnsProvider.Hetzner = &entity.AcmeDnsProviderHetzner{
				APIToken: entity.NewEncryptedField(req.Hetzner.APIToken),
			}
		}
	case base.AcmeDnsProviderHuaweiCloud:
		if req.HuaweiCloud != nil {
			acmeDnsProvider.HuaweiCloud = &entity.AcmeDnsProviderHuaweiCloud{
				AccessKey: req.HuaweiCloud.AccessKey,
				SecretKey: entity.NewEncryptedField(req.HuaweiCloud.SecretKey),
				Region:    req.HuaweiCloud.Region,
			}
		}
	case base.AcmeDnsProviderNamecheap:
		if req.Namecheap != nil {
			acmeDnsProvider.Namecheap = &entity.AcmeDnsProviderNamecheap{
				APIUser: req.Namecheap.APIUser,
				APIKey:  entity.NewEncryptedField(req.Namecheap.APIKey),
			}
		}
	case base.AcmeDnsProviderRFC2136:
		if req.RFC2136 != nil {
			acmeDnsProvider.RFC2136 = &entity.AcmeDnsProviderRFC2136{
				Nameserver:    req.RFC2136.Nameserver,
				TSIGKeyName:   req.RFC2136.TSIGKeyName,
				TSIGSecret:    entity.NewEncryptedField(req.RFC2136.TSIGSecret),
				TSIGAlgorithm: req.RFC2136.TSIGAlgorithm,
			}
		}
	case base.AcmeDnsProviderRoute53:
		if req.Route53 != nil {
			acmeDnsProvider.Route53 = &entity.AcmeDnsProviderRoute53{
				AccessKeyID:     req.Route53.AccessKeyID,
				SecretAccessKey: entity.NewEncryptedField(req.Route53.SecretAccessKey),
				HostedZoneID:    req.Route53.HostedZoneID,
				Region:          req.Route53.Region,
			}
		}
	case base.AcmeDnsProviderTencentCloud:
		if req.TencentCloud != nil {
			acmeDnsProvider.TencentCloud = &entity.AcmeDnsProviderTencentCloud{
				SecretID:  req.TencentCloud.SecretID,
				SecretKey: entity.NewEncryptedField(req.TencentCloud.SecretKey),
				Region:    req.TencentCloud.Region,
			}
		}
	}
	return acmeDnsProvider
}

type AcmeDnsProviderAcmeDNSReq struct {
	APIBase        string   `json:"apiBase"`
	AllowList      []string `json:"allowList"`
	StoragePath    string   `json:"storagePath"`
	StorageBaseURL string   `json:"storageBaseUrl"`
}

func (req *AcmeDnsProviderAcmeDNSReq) validate(field string) (res []vld.Validator) {
	if req == nil {
		return nil
	}
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStr(&req.APIBase, true, 1, len1000, field+"apiBase")...)
	res = append(res, basedto.ValidateStr(&req.StoragePath, false, 1, len1000, field+"storagePath")...)
	res = append(res, basedto.ValidateStr(&req.StorageBaseURL, false, 1, len1000, field+"storageBaseUrl")...)
	return res
}

type AcmeDnsProviderAzureReq struct {
	ClientID          string `json:"clientId"`
	ClientSecret      string `json:"clientSecret"`
	SubscriptionID    string `json:"subscriptionId"`
	TenantID          string `json:"tenantId"`
	ResourceGroupName string `json:"resourceGroupName"`
}

func (req *AcmeDnsProviderAzureReq) validate(field string) (res []vld.Validator) {
	if req == nil {
		return nil
	}
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStr(&req.ClientID, true, 1, len200, field+"clientId")...)
	res = append(res, basedto.ValidateStr(&req.ClientSecret, true, 1, len1000, field+"clientSecret")...)
	res = append(res, basedto.ValidateStr(&req.SubscriptionID, false, 1, len200, field+"subscriptionId")...)
	res = append(res, basedto.ValidateStr(&req.TenantID, false, 1, len200, field+"tenantId")...)
	res = append(res, basedto.ValidateStr(&req.ResourceGroupName, false, 1, len200, field+"resourceGroupName")...)
	return res
}

type AcmeDnsProviderBaiduCloudReq struct {
	AccessKey string `json:"accessKey"`
	SecretKey string `json:"secretKey"`
}

func (req *AcmeDnsProviderBaiduCloudReq) validate(field string) (res []vld.Validator) {
	if req == nil {
		return nil
	}
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStr(&req.AccessKey, true, 1, len200, field+"accessKey")...)
	res = append(res, basedto.ValidateStr(&req.SecretKey, true, 1, len1000, field+"secretKey")...)
	return res
}

type AcmeDnsProviderCloudflareReq struct {
	AuthToken string `json:"authToken"`
}

func (req *AcmeDnsProviderCloudflareReq) validate(field string) (res []vld.Validator) {
	if req == nil {
		return nil
	}
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStr(&req.AuthToken, true, 1, len1000, field+"authToken")...)
	return res
}

type AcmeDnsProviderDigitalOceanReq struct {
	AuthToken string `json:"authToken"`
}

func (req *AcmeDnsProviderDigitalOceanReq) validate(field string) (res []vld.Validator) {
	if req == nil {
		return nil
	}
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStr(&req.AuthToken, true, 1, len1000, field+"authToken")...)
	return res
}

type AcmeDnsProviderGCloudReq struct {
	ProjectID      string `json:"projectId"`
	ServiceAccount string `json:"serviceAccount"`
}

func (req *AcmeDnsProviderGCloudReq) validate(field string) (res []vld.Validator) {
	if req == nil {
		return nil
	}
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStr(&req.ServiceAccount, true, 1, len10000, field+"serviceAccount")...)
	res = append(res, basedto.ValidateStr(&req.ProjectID, false, 1, len200, field+"projectId")...)
	return res
}

type AcmeDnsProviderGoDaddyReq struct {
	APIKey    string `json:"apiKey"`
	APISecret string `json:"apiSecret"`
}

func (req *AcmeDnsProviderGoDaddyReq) validate(field string) (res []vld.Validator) {
	if req == nil {
		return nil
	}
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStr(&req.APIKey, true, 1, len200, field+"apiKey")...)
	res = append(res, basedto.ValidateStr(&req.APISecret, true, 1, len1000, field+"apiSecret")...)
	return res
}

type AcmeDnsProviderHetznerReq struct {
	APIToken string `json:"apiToken"`
}

func (req *AcmeDnsProviderHetznerReq) validate(field string) (res []vld.Validator) {
	if req == nil {
		return nil
	}
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStr(&req.APIToken, true, 1, len1000, field+"apiToken")...)
	return res
}

type AcmeDnsProviderHuaweiCloudReq struct {
	AccessKey string `json:"accessKey"`
	SecretKey string `json:"secretKey"`
	Region    string `json:"region,omitempty"`
}

func (req *AcmeDnsProviderHuaweiCloudReq) validate(field string) (res []vld.Validator) {
	if req == nil {
		return nil
	}
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStr(&req.AccessKey, true, 1, len200, field+"accessKey")...)
	res = append(res, basedto.ValidateStr(&req.SecretKey, true, 1, len1000, field+"secretKey")...)
	res = append(res, basedto.ValidateStr(&req.Region, false, 0, len200, field+"region")...)
	return res
}

type AcmeDnsProviderNamecheapReq struct {
	APIUser string `json:"apiUser"`
	APIKey  string `json:"apiKey"`
}

func (req *AcmeDnsProviderNamecheapReq) validate(field string) (res []vld.Validator) {
	if req == nil {
		return nil
	}
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStr(&req.APIUser, true, 1, len200, field+"apiUser")...)
	res = append(res, basedto.ValidateStr(&req.APIKey, true, 1, len1000, field+"apiKey")...)
	return res
}

type AcmeDnsProviderRFC2136Req struct {
	Nameserver    string `json:"nameserver"`
	TSIGKeyName   string `json:"tsigKeyName"`
	TSIGSecret    string `json:"tsigSecret"`
	TSIGAlgorithm string `json:"tsigAlgorithm"`
}

func (req *AcmeDnsProviderRFC2136Req) validate(field string) (res []vld.Validator) {
	if req == nil {
		return nil
	}
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStr(&req.Nameserver, true, 1, len200, field+"nameserver")...)
	res = append(res, basedto.ValidateStr(&req.TSIGKeyName, true, 1, len200, field+"tsigKeyName")...)
	res = append(res, basedto.ValidateStr(&req.TSIGSecret, true, 1, len1000, field+"tsigSecret")...)
	res = append(res, basedto.ValidateStr(&req.TSIGAlgorithm, false, 1, len200, field+"tsigAlgorithm")...)
	return res
}

type AcmeDnsProviderRoute53Req struct {
	AccessKeyID     string `json:"accessKeyId"`
	SecretAccessKey string `json:"secretAccessKey"`
	HostedZoneID    string `json:"hostedZoneId,omitempty"`
	Region          string `json:"region,omitempty"`
}

func (req *AcmeDnsProviderRoute53Req) validate(field string) (res []vld.Validator) {
	if req == nil {
		return nil
	}
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStr(&req.AccessKeyID, true, 1, len200, field+"accessKeyId")...)
	res = append(res, basedto.ValidateStr(&req.SecretAccessKey, true, 1, len1000, field+"secretAccessKey")...)
	res = append(res, basedto.ValidateStr(&req.HostedZoneID, false, 0, len200, field+"hostedZoneId")...)
	res = append(res, basedto.ValidateStr(&req.Region, false, 0, len200, field+"region")...)
	return res
}

type AcmeDnsProviderTencentCloudReq struct {
	SecretID  string `json:"secretId"`
	SecretKey string `json:"secretKey"`
	Region    string `json:"region,omitempty"`
}

func (req *AcmeDnsProviderTencentCloudReq) validate(field string) (res []vld.Validator) {
	if req == nil {
		return nil
	}
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStr(&req.SecretID, true, 1, len200, field+"secretId")...)
	res = append(res, basedto.ValidateStr(&req.SecretKey, true, 1, len1000, field+"secretKey")...)
	res = append(res, basedto.ValidateStr(&req.Region, false, 0, len200, field+"region")...)
	return res
}

func (req *AcmeDnsProviderBaseReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStr(&req.Name, true, 1, len200, field+"name")...)
	res = append(res, basedto.ValidateStrIn(&req.Kind, true, base.AllAcmeDnsProviders, field+"kind")...)

	switch req.Kind {
	case base.AcmeDnsProviderAcmeDNS:
		res = append(res, basedto.ValidateCond(req.AcmeDNS != nil, field+"acmeDns")...)
		res = append(res, req.AcmeDNS.validate(field+"acmeDns")...)
	case base.AcmeDnsProviderAzure:
		res = append(res, basedto.ValidateCond(req.Azure != nil, field+"azure")...)
		res = append(res, req.Azure.validate(field+"azure")...)
	case base.AcmeDnsProviderBaiduCloud:
		res = append(res, basedto.ValidateCond(req.BaiduCloud != nil, field+"baiduCloud")...)
		res = append(res, req.BaiduCloud.validate(field+"baiduCloud")...)
	case base.AcmeDnsProviderCloudflare:
		res = append(res, basedto.ValidateCond(req.Cloudflare != nil, field+"cloudflare")...)
		res = append(res, req.Cloudflare.validate(field+"cloudflare")...)
	case base.AcmeDnsProviderDigitalOcean:
		res = append(res, basedto.ValidateCond(req.DigitalOcean != nil, field+"digitalOcean")...)
		res = append(res, req.DigitalOcean.validate(field+"digitalOcean")...)
	case base.AcmeDnsProviderGCloud:
		res = append(res, basedto.ValidateCond(req.GCloud != nil, field+"gCloud")...)
		res = append(res, req.GCloud.validate(field+"gCloud")...)
	case base.AcmeDnsProviderGoDaddy:
		res = append(res, basedto.ValidateCond(req.GoDaddy != nil, field+"goDaddy")...)
		res = append(res, req.GoDaddy.validate(field+"goDaddy")...)
	case base.AcmeDnsProviderHetzner:
		res = append(res, basedto.ValidateCond(req.Hetzner != nil, field+"hetzner")...)
		res = append(res, req.Hetzner.validate(field+"hetzner")...)
	case base.AcmeDnsProviderHuaweiCloud:
		res = append(res, basedto.ValidateCond(req.HuaweiCloud != nil, field+"huaweiCloud")...)
		res = append(res, req.HuaweiCloud.validate(field+"huaweiCloud")...)
	case base.AcmeDnsProviderNamecheap:
		res = append(res, basedto.ValidateCond(req.Namecheap != nil, field+"namecheap")...)
		res = append(res, req.Namecheap.validate(field+"namecheap")...)
	case base.AcmeDnsProviderRFC2136:
		res = append(res, basedto.ValidateCond(req.RFC2136 != nil, field+"rfc2136")...)
		res = append(res, req.RFC2136.validate(field+"rfc2136")...)
	case base.AcmeDnsProviderRoute53:
		res = append(res, basedto.ValidateCond(req.Route53 != nil, field+"route53")...)
		res = append(res, req.Route53.validate(field+"route53")...)
	case base.AcmeDnsProviderTencentCloud:
		res = append(res, basedto.ValidateCond(req.TencentCloud != nil, field+"tencentCloud")...)
		res = append(res, req.TencentCloud.validate(field+"tencentCloud")...)
	}
	return res
}

func NewCreateAcmeDnsProviderReq() *CreateAcmeDnsProviderReq {
	return &CreateAcmeDnsProviderReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *CreateAcmeDnsProviderReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type CreateAcmeDnsProviderResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *basedto.ObjectIDResp `json:"data"`
}
