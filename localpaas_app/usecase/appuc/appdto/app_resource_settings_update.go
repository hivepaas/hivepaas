package appdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
)

type UpdateAppResourceSettingsReq struct {
	ProjectID string `json:"-"`
	AppID     string `json:"-"`

	Reservations *ResourceReservations `json:"reservations"`
	Limits       *ResourceLimits       `json:"limits"`
	Ulimits      []*Ulimit             `json:"ulimits"`
	Capabilities *Capabilities         `json:"capabilities"`

	UpdateVer int `json:"updateVer"`
}

func NewUpdateAppResourceSettingsReq() *UpdateAppResourceSettingsReq {
	return &UpdateAppResourceSettingsReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateAppResourceSettingsReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectID")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appID")...)
	// TODO: validate env var input
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateAppResourceSettingsResp struct {
	Meta *basedto.Meta                      `json:"meta"`
	Data *UpdateAppResourceSettingsDataResp `json:"data"`
}

type UpdateAppResourceSettingsDataResp struct {
}
