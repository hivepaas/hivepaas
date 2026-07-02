package appsettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
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
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	// TODO: add validation
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateAppStorageSettingsResp struct {
	Meta *basedto.Meta                     `json:"meta"`
	Data *UpdateAppStorageSettingsDataResp `json:"data"`
}

type UpdateAppStorageSettingsDataResp struct {
}
