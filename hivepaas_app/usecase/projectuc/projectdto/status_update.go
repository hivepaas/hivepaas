package projectdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type UpdateProjectStatusReq struct {
	ID        string             `json:"-"`
	UpdateVer int                `json:"updateVer"`
	Status    base.ProjectStatus `json:"status"`
}

func NewUpdateProjectStatusReq() *UpdateProjectStatusReq {
	return &UpdateProjectStatusReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateProjectStatusReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateStrIn(&req.Status, true, base.AllProjectStatuses, "status")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateProjectStatusResp struct {
	Meta *basedto.Meta `json:"meta"`
}
