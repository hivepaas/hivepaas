package domainsettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/copier"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type GetDomainSettingsReq struct {
	settings.GetUniqueSettingReq
}

func NewGetDomainSettingsReq() *GetDomainSettingsReq {
	return &GetDomainSettingsReq{}
}

func (req *GetDomainSettingsReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.GetUniqueSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetDomainSettingsResp struct {
	Meta *basedto.Meta       `json:"meta"`
	Data *DomainSettingsResp `json:"data"`
}

type DomainSettingsResp struct {
	*settings.BaseSettingResp
	RootDomain     string                  `json:"rootDomain"`
	AllowedDomains []string                `json:"allowedDomains"`
	CertSettings   *DomainCertSettingsResp `json:"certSettings"`
}

type DomainCertSettingsResp struct {
	CertType    base.SSLCertType  `json:"certType"`
	KeyType     base.SSLKeyType   `json:"keyType"`
	ValidPeriod timeutil.Duration `json:"validPeriod,omitempty"`
	Email       string            `json:"email"`
	AutoRenew   bool              `json:"autoRenew,omitempty"`
}

func TransformDomainSettings(
	setting *entity.Setting,
	_ *entity.RefObjects,
) (resp *DomainSettingsResp, err error) {
	config := setting.MustAsDomainSettings()
	if err = copier.Copy(&resp, config); err != nil {
		return nil, apperrors.New(err)
	}

	resp.BaseSettingResp, err = settings.TransformSettingBase(setting)
	if err != nil {
		return nil, apperrors.New(err)
	}

	return resp, nil
}
