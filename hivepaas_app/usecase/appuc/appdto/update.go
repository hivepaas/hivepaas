package appdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
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
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, req.validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

func (req *UpdateAppReq) ModifyRequest() error {
	return req.modifyRequest()
}

type UpdateAppResp struct {
	Meta *basedto.Meta `json:"meta"`
}
