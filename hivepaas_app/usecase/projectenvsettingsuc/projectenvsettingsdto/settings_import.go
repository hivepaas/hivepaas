package projectenvsettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
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
func (req *ImportSettingsToProjectEnvReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
	validators = append(validators, basedto.ValidateObjectIDSliceReq(req.Settings, true, 1, "settings")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type ImportSettingsToProjectEnvResp struct {
	Meta *basedto.Meta `json:"meta"`
}
