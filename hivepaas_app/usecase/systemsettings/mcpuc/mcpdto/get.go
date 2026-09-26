package mcpdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type GetMCPSettingsReq struct {
	settings.GetUniqueSettingReq
}

func NewGetMCPSettingsReq() *GetMCPSettingsReq {
	return &GetMCPSettingsReq{}
}

func (req *GetMCPSettingsReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 5) //nolint:mnd
	validators = append(validators, req.GetUniqueSettingReq.Validate()...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetMCPSettingsResp struct {
	Meta *basedto.Meta    `json:"meta"`
	Data *MCPSettingsResp `json:"data"`
}

// MCPSettingsResp is whether HivePaaS serves the Model Context Protocol. A
// setting never saved reads as off, with updateVer 0.
type MCPSettingsResp struct {
	Enabled    bool `json:"enabled"`
	AllowWrite bool `json:"allowWrite"`
	UpdateVer  int  `json:"updateVer"`
}

func TransformMCPSettings(setting *entity.Setting) (*MCPSettingsResp, error) {
	data, err := setting.AsMCPSettings()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &MCPSettingsResp{Enabled: data.Enabled, AllowWrite: data.AllowWrite, UpdateVer: setting.UpdateVer}, nil
}
