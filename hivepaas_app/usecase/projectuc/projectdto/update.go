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
	validators := make([]vld.Validator, 0, 5) //nolint:mnd
	validators = append(validators, req.validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

func (req *UpdateProjectReq) ModifyRequest() error {
	return req.modifyRequest()
}

type UpdateProjectResp struct {
	Meta *basedto.Meta `json:"meta"`
}
