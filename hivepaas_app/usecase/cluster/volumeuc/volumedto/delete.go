package volumedto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type DeleteVolumeReq struct {
	settings.DeleteSettingReq
}

func NewDeleteVolumeReq() *DeleteVolumeReq {
	return &DeleteVolumeReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *DeleteVolumeReq) Validate() apperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 5) //nolint:mnd
	validators = append(validators, req.DeleteSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type DeleteVolumeResp struct {
	Meta *basedto.Meta `json:"meta"`
}
