package appdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
)

type UpdateAppServiceSettingsReq struct {
	ProjectID string `json:"-"`
	AppID     string `json:"-"`

	ModeSpec  *ServiceModeSpec `json:"modeSpec"`
	Placement *Placement       `json:"placement"`

	UpdateVer int `json:"updateVer"`
}

func NewUpdateAppServiceSettingsReq() *UpdateAppServiceSettingsReq {
	return &UpdateAppServiceSettingsReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateAppServiceSettingsReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectID")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appID")...)
	// TODO: validate env var input
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateAppServiceSettingsResp struct {
	Meta *basedto.Meta                     `json:"meta"`
	Data *UpdateAppServiceSettingsDataResp `json:"data"`
}

type UpdateAppServiceSettingsDataResp struct {
}
