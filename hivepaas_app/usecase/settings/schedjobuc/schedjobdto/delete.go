package schedjobdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type DeleteSchedJobReq struct {
	settings.DeleteSettingReq
}

func NewDeleteSchedJobReq() *DeleteSchedJobReq {
	return &DeleteSchedJobReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *DeleteSchedJobReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.DeleteSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type DeleteSchedJobResp struct {
	Meta *basedto.Meta `json:"meta"`
}
