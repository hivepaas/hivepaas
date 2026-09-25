package settingmountdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type UpdateSettingMountReq struct {
	settings.UpdateSettingReq
	*SettingMountBaseReq
}

func NewUpdateSettingMountReq() *UpdateSettingMountReq {
	return &UpdateSettingMountReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateSettingMountReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.UpdateSettingReq.Validate()...)
	if req.SettingMountBaseReq == nil {
		req.SettingMountBaseReq = &SettingMountBaseReq{}
	}
	validators = append(validators, req.validate()...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateSettingMountResp struct {
	Meta *basedto.Meta `json:"meta"`
}
