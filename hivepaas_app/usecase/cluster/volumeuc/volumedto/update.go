package volumedto

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type UpdateVolumeReq struct {
	settings.UpdateSettingReq
}

func NewUpdateVolumeReq() *UpdateVolumeReq {
	return &UpdateVolumeReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateVolumeReq) Validate() apperrors.ValidationErrors {
	return nil
}

type UpdateVolumeResp struct {
	Meta *basedto.Meta `json:"meta"`
}
