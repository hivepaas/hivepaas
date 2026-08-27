package commandpipedto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type DeleteCommandPipeReq struct {
	settings.DeleteSettingReq
}

func NewDeleteCommandPipeReq() *DeleteCommandPipeReq {
	return &DeleteCommandPipeReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *DeleteCommandPipeReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.DeleteSettingReq.Validate()...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type DeleteCommandPipeResp struct {
	Meta *basedto.Meta `json:"meta"`
}
