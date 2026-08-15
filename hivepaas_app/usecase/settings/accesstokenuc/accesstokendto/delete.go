package accesstokendto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type DeleteAccessTokenReq struct {
	settings.DeleteSettingReq
}

func NewDeleteAccessTokenReq() *DeleteAccessTokenReq {
	return &DeleteAccessTokenReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *DeleteAccessTokenReq) Validate() apperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 5) //nolint:mnd
	validators = append(validators, req.DeleteSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type DeleteAccessTokenResp struct {
	Meta *basedto.Meta `json:"meta"`
}
