package appdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
)

type UpdateAppReq struct {
	ID        string `json:"-"`
	ProjectID string `json:"-"`
	UpdateVer int    `json:"updateVer"`
	*AppBaseReq
}

func NewUpdateAppReq() *UpdateAppReq {
	return &UpdateAppReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateAppReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ID, true, "id")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectID")...)
	validators = append(validators, req.validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateAppResp struct {
	Meta *basedto.Meta `json:"meta"`
}
