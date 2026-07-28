package projectsettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type ImportSettingsToProjectReq struct {
	ProjectID    string                   `json:"-"`
	ProjectEnvID string                   `json:"-"`
	Settings     basedto.ObjectIDSliceReq `json:"settings"`
	CanViewData  bool                     `json:"canViewData"`
}

func NewImportSettingsToProjectReq() *ImportSettingsToProjectReq {
	return &ImportSettingsToProjectReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *ImportSettingsToProjectReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateStr(&req.ProjectEnvID, false,
		base.ProjectEnvMinLen, base.ProjectEnvMaxLen, "projectEnv")...)
	validators = append(validators, basedto.ValidateObjectIDSliceReq(req.Settings, true, 1, "settings")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type ImportSettingsToProjectResp struct {
	Meta *basedto.Meta `json:"meta"`
}
