package projectdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type UpdateProjectReq struct {
	ID        string `json:"-"`
	UpdateVer int    `json:"updateVer"`
	*ProjectBaseReq
}

func NewUpdateProjectReq() *UpdateProjectReq {
	return &UpdateProjectReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateProjectReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateProjectResp struct {
	Meta *basedto.Meta `json:"meta"`
}
