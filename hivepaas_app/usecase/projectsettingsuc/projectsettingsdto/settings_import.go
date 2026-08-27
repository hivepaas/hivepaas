package projectsettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type ImportSettingsToProjectReq struct {
	ProjectID   string                   `json:"-"`
	Settings    basedto.ObjectIDSliceReq `json:"settings"`
	CanViewData bool                     `json:"canViewData"`
}

func NewImportSettingsToProjectReq() *ImportSettingsToProjectReq {
	return &ImportSettingsToProjectReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *ImportSettingsToProjectReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateObjectIDSliceReq(req.Settings, true, 1, "settings")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type ImportSettingsToProjectResp struct {
	Meta *basedto.Meta `json:"meta"`
}
