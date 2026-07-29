package projectenvsettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type ImportSettingsToProjectEnvReq struct {
	ProjectID    string                   `json:"-"`
	ProjectEnvID string                   `json:"-"`
	Settings     basedto.ObjectIDSliceReq `json:"settings"`
	CanViewData  bool                     `json:"canViewData"`
}

func NewImportSettingsToProjectEnvReq() *ImportSettingsToProjectEnvReq {
	return &ImportSettingsToProjectEnvReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *ImportSettingsToProjectEnvReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateStr(&req.ProjectEnvID, true,
		base.ProjectEnvMinLen, base.ProjectEnvMaxLen, "projectEnv")...)
	validators = append(validators, basedto.ValidateObjectIDSliceReq(req.Settings, true, 1, "settings")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type ImportSettingsToProjectEnvResp struct {
	Meta *basedto.Meta `json:"meta"`
}
