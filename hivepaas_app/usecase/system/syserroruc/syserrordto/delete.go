package syserrordto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type DeleteSysErrorReq struct {
	ID string `json:"-"`
}

func NewDeleteSysErrorReq() *DeleteSysErrorReq {
	return &DeleteSysErrorReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *DeleteSysErrorReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ID, true, "id")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type DeleteSysErrorResp struct {
	Meta *basedto.Meta `json:"meta"`
}
