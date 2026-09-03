package domainsettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

const (
	allowedDomainsMaxLen = 100
)

type UpdateDomainSettingsReq struct {
	settings.UpdateUniqueSettingReq
	*DomainSettingsBaseReq
}

type DomainSettingsBaseReq struct {
	RootDomain     string                 `json:"rootDomain"`
	AllowedDomains []string               `json:"allowedDomains"`
	CertSettings   *DomainCertSettingsReq `json:"certSettings"`
}

func (req *DomainSettingsBaseReq) ToEntity() *entity.DomainSettings {
	return &entity.DomainSettings{
		RootDomain:     req.RootDomain,
		AllowedDomains: req.AllowedDomains,
		CertSettings:   req.CertSettings.ToEntity(),
	}
}

func (req *DomainSettingsBaseReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateSliceEx(req.AllowedDomains, true, 0, allowedDomainsMaxLen,
		nil, field+"allowedDomains")...)
	res = append(res, req.CertSettings.validate(field+"certSettings")...)
	return res
}

type DomainCertSettingsReq struct {
	CertType    base.SSLCertType  `json:"certType"`
	KeyType     base.SSLKeyType   `json:"keyType"`
	ValidPeriod timeutil.Duration `json:"validPeriod,omitempty"`
	Email       string            `json:"email"`
	AutoRenew   bool              `json:"autoRenew,omitempty"`
}

func (req *DomainCertSettingsReq) ToEntity() *entity.DomainCertSettings {
	if req == nil {
		return nil
	}
	return &entity.DomainCertSettings{
		CertType:    req.CertType,
		KeyType:     req.KeyType,
		ValidPeriod: req.ValidPeriod,
		Email:       req.Email,
		AutoRenew:   req.AutoRenew,
	}
}

func (req *DomainCertSettingsReq) validate(field string) (res []vld.Validator) {
	if req == nil {
		return nil
	}
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStrIn(&req.CertType, false, base.AllSSLCertTypes, field+"certType")...)
	res = append(res, basedto.ValidateStrIn(&req.KeyType, false, base.AllSSLKeyTypes, field+"keyType")...)
	return res
}

func NewUpdateDomainSettingsReq() *UpdateDomainSettingsReq {
	return &UpdateDomainSettingsReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateDomainSettingsReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.UpdateUniqueSettingReq.Validate()...)
	validators = append(validators, req.validate("")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateDomainSettingsResp struct {
	Meta *basedto.Meta `json:"meta"`
}
