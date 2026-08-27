package imservicedto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type DeleteIMServiceReq struct {
	settings.DeleteSettingReq
}

func NewDeleteIMServiceReq() *DeleteIMServiceReq {
	return &DeleteIMServiceReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *DeleteIMServiceReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.DeleteSettingReq.Validate()...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type DeleteIMServiceResp struct {
	Meta *basedto.Meta `json:"meta"`
}
