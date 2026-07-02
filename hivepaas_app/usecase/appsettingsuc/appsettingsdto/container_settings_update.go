package appsettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type UpdateAppContainerSettingsReq struct {
	ProjectID string `json:"-"`
	AppID     string `json:"-"`

	*BaseContainerSettings
	UpdateVer int `json:"updateVer"`
}

func NewUpdateAppContainerSettingsReq() *UpdateAppContainerSettingsReq {
	return &UpdateAppContainerSettingsReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateAppContainerSettingsReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	// TODO: add validation
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateAppContainerSettingsResp struct {
	Meta *basedto.Meta `json:"meta"`
}
