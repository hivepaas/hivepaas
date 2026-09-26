package mcpdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type UpdateMCPSettingsReq struct {
	settings.UpdateUniqueSettingReq
	Enabled    bool `json:"enabled"`
	AllowWrite bool `json:"allowWrite"`
}

func NewUpdateMCPSettingsReq() *UpdateMCPSettingsReq {
	return &UpdateMCPSettingsReq{}
}

func (req *UpdateMCPSettingsReq) ToEntity() *entity.MCPSettings {
	return &entity.MCPSettings{Enabled: req.Enabled, AllowWrite: req.AllowWrite}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateMCPSettingsReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 5) //nolint:mnd
	validators = append(validators, req.UpdateUniqueSettingReq.Validate()...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateMCPSettingsResp struct {
	Meta *basedto.Meta `json:"meta"`
}
