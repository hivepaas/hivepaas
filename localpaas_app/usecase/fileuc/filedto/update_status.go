package filedto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
)

type UpdateFileStatusReq struct {
	ID     string           `json:"-"`
	Status *base.FileStatus `json:"status"`
}

func NewUpdateFileStatusReq() *UpdateFileStatusReq {
	return &UpdateFileStatusReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateFileStatusReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateFileStatusResp struct {
	Meta *basedto.Meta `json:"meta"`
}
