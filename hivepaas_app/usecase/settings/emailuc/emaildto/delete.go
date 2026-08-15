package emaildto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type DeleteEmailReq struct {
	settings.DeleteSettingReq
}

func NewDeleteEmailReq() *DeleteEmailReq {
	return &DeleteEmailReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *DeleteEmailReq) Validate() apperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 5) //nolint:mnd
	validators = append(validators, req.DeleteSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type DeleteEmailResp struct {
	Meta *basedto.Meta `json:"meta"`
}
