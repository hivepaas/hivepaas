package appdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
)

type UpdateAppContainerSettingsReq struct {
	ProjectID string `json:"-"`
	AppID     string `json:"-"`

	*ContainerSpec

	UpdateVer int `json:"updateVer"`
}

func NewUpdateAppContainerSettingsReq() *UpdateAppContainerSettingsReq {
	return &UpdateAppContainerSettingsReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateAppContainerSettingsReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectID")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appID")...)
	// TODO: validate env var input
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateAppContainerSettingsResp struct {
	Meta *basedto.Meta                       `json:"meta"`
	Data *UpdateAppContainerSettingsDataResp `json:"data"`
}

type UpdateAppContainerSettingsDataResp struct {
}
