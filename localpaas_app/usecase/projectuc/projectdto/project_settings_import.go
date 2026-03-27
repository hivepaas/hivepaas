package projectdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
)

type ImportSettingsToProjectReq struct {
	ProjectID       string                   `json:"-"`
	Settings        basedto.ObjectIDSliceReq `json:"settings"`
	DataViewAllowed bool                     `json:"dataViewAllowed"`
}

func NewImportSettingsToProjectReq() *ImportSettingsToProjectReq {
	return &ImportSettingsToProjectReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *ImportSettingsToProjectReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectID")...)
	validators = append(validators, basedto.ValidateObjectIDSliceReq(req.Settings, true, 1, "settings")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type ImportSettingsToProjectResp struct {
	Meta *basedto.Meta `json:"meta"`
}
