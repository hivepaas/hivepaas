package appdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
)

type UpdateAppTokenReq struct {
	ID        string `json:"-"`
	ProjectID string `json:"-"`
	UpdateVer int    `json:"updateVer"`
}

func NewUpdateAppTokenReq() *UpdateAppTokenReq {
	return &UpdateAppTokenReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateAppTokenReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ID, true, "id")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectID")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateAppTokenResp struct {
	Meta *basedto.Meta     `json:"meta"`
	Data *AppTokenDataResp `json:"data"`
}

type AppTokenDataResp struct {
	Token string `json:"token"`
}
