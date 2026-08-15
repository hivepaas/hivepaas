package commandpipedto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type UpdateCommandPipeReq struct {
	settings.UpdateSettingReq
	*CommandPipeBaseReq
}

func NewUpdateCommandPipeReq() *UpdateCommandPipeReq {
	return &UpdateCommandPipeReq{}
}

func (req *UpdateCommandPipeReq) ModifyRequest() error {
	return req.modifyRequest()
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateCommandPipeReq) Validate() apperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 5) //nolint:mnd
	validators = append(validators, req.validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateCommandPipeResp struct {
	Meta *basedto.Meta `json:"meta"`
}
