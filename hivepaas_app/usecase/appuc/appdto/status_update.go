package appdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type UpdateAppStatusReq struct {
	ID        string         `json:"-"`
	ProjectID string         `json:"-"`
	UpdateVer int            `json:"updateVer"`
	Status    base.AppStatus `json:"status"`
}

func NewUpdateAppStatusReq() *UpdateAppStatusReq {
	return &UpdateAppStatusReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateAppStatusReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ID, true, "id")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateStrIn(&req.Status, true, base.AllAppStatuses, "status")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateAppStatusResp struct {
	Meta *basedto.Meta `json:"meta"`
}
