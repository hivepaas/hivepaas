package appsettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type UpdateAppResourceSettingsReq struct {
	ProjectID string `json:"-"`
	AppID     string `json:"-"`

	Reservations *ResourceReservations `json:"reservations"`
	Limits       *ResourceLimits       `json:"limits"`
	Memory       *Memory               `json:"memory"`
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
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	// TODO: add validation
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateAppResourceSettingsResp struct {
	Meta *basedto.Meta                      `json:"meta"`
	Data *UpdateAppResourceSettingsDataResp `json:"data"`
}

type UpdateAppResourceSettingsDataResp struct {
}
