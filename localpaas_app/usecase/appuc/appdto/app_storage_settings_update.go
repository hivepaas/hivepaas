package appdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
)

type UpdateAppStorageSettingsReq struct {
	ProjectID string `json:"-"`
	AppID     string `json:"-"`

	Mounts []*Mount `json:"mounts"`

	UpdateVer int `json:"updateVer"`
}

func NewUpdateAppStorageSettingsReq() *UpdateAppStorageSettingsReq {
	return &UpdateAppStorageSettingsReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateAppStorageSettingsReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectID")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appID")...)
	// TODO: validate env var input
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateAppStorageSettingsResp struct {
	Meta *basedto.Meta                     `json:"meta"`
	Data *UpdateAppStorageSettingsDataResp `json:"data"`
}

type UpdateAppStorageSettingsDataResp struct {
}
