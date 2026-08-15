package userdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type DeleteUserReq struct {
	ID string `json:"-"`
}

func NewDeleteUserReq() *DeleteUserReq {
	return &DeleteUserReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *DeleteUserReq) Validate() apperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 5) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ID, true, "id")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type DeleteUserResp struct {
	Meta *basedto.Meta `json:"meta"`
}
