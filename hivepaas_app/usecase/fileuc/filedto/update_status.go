package filedto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
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
	validators = append(validators, basedto.ValidateID(&req.ID, true, "id")...)
	// TODO: add validation
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateFileStatusResp struct {
	Meta *basedto.Meta `json:"meta"`
}
