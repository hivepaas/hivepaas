package periodicjobdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type DeletePeriodicJobReq struct {
	settings.DeleteSettingReq
}

func NewDeletePeriodicJobReq() *DeletePeriodicJobReq {
	return &DeletePeriodicJobReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *DeletePeriodicJobReq) Validate() apperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.DeleteSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type DeletePeriodicJobResp struct {
	Meta *basedto.Meta `json:"meta"`
}
