package vulnignoredto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type GetVulnIgnoreReq struct {
	settings.GetSettingReq
}

func NewGetVulnIgnoreReq() *GetVulnIgnoreReq {
	return &GetVulnIgnoreReq{}
}

func (req *GetVulnIgnoreReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.GetSettingReq.Validate()...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetVulnIgnoreResp struct {
	Meta *basedto.Meta   `json:"meta"`
	Data *VulnIgnoreResp `json:"data"`
}

// VulnIgnoreResp is an ignore: what it sets aside and why. When it stops
// applying is the setting's expireAt.
type VulnIgnoreResp struct {
	*settings.BaseSettingResp
	Scope   string `json:"scope"`
	VulnID  string `json:"vulnId"`
	Package string `json:"package,omitempty"`
	Reason  string `json:"reason"`
	SetBy   string `json:"setBy,omitempty"`
}

func TransformVulnIgnore(
	setting *entity.Setting,
	_ *entity.RefObjects,
) (resp *VulnIgnoreResp, err error) {
	ignore, err := setting.AsVulnerabilityIgnore()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	resp = &VulnIgnoreResp{
		Scope:   scopeName(setting),
		VulnID:  ignore.VulnID,
		Package: ignore.Package,
		Reason:  ignore.Reason,
		SetBy:   ignore.SetBy,
	}
	resp.BaseSettingResp, err = settings.TransformSettingBase(setting)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return resp, nil
}

// scopeName is where the ignore was set, as the screens say it: global for the
// installation's.
func scopeName(setting *entity.Setting) string {
	if setting.Scope == "" {
		return "global"
	}
	return string(setting.Scope)
}
