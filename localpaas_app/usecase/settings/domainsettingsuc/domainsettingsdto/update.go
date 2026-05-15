package domainsettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
)

type UpdateDomainSettingsReq struct {
	settings.UpdateUniqueSettingReq
	*DomainSettingsBaseReq
}

type DomainSettingsBaseReq struct {
	RootDomain   string                 `json:"rootDomain"`
	CertSettings *DomainCertSettingsReq `json:"certSettings"`
}

func (req *DomainSettingsBaseReq) ToEntity() *entity.DomainSettings {
	return &entity.DomainSettings{
		RootDomain:   req.RootDomain,
		CertSettings: req.CertSettings.ToEntity(),
	}
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

// nolint
func (req *DomainSettingsBaseReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	// TODO: add validation
	return res
}

func NewUpdateDomainSettingsReq() *UpdateDomainSettingsReq {
	return &UpdateDomainSettingsReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateDomainSettingsReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateDomainSettingsResp struct {
	Meta *basedto.Meta `json:"meta"`
}
