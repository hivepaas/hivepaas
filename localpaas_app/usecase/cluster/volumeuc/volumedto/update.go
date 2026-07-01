package volumedto

import (
	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
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
