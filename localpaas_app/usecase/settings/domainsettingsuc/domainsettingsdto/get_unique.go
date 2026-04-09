package domainsettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/copier"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
)

type GetUniqueDomainSettingsReq struct {
	settings.GetUniqueSettingReq
}

func NewGetUniqueDomainSettingsReq() *GetUniqueDomainSettingsReq {
	return &GetUniqueDomainSettingsReq{}
}

func (req *GetUniqueDomainSettingsReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.GetUniqueSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetUniqueDomainSettingsResp struct {
	Meta *basedto.Meta       `json:"meta"`
	Data *DomainSettingsResp `json:"data"`
}

type DomainSettingsResp struct {
	*settings.BaseSettingResp
	RootDomain   string                  `json:"rootDomain"`
	CertSettings *DomainCertSettingsResp `json:"certSettings"`
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
		return nil, apperrors.Wrap(err)
	}

	resp.BaseSettingResp, err = settings.TransformSettingBase(setting)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return resp, nil
}
