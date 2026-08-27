package appdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type UpdateAppStatusReq struct {
	ProjectID    string         `json:"-"`
	ProjectEnvID string         `json:"-"`
	AppID        string         `json:"-"`
	UpdateVer    int            `json:"updateVer"`
	Status       base.AppStatus `json:"status"`
}

func NewUpdateAppStatusReq() *UpdateAppStatusReq {
	return &UpdateAppStatusReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateAppStatusReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	validators = append(validators, basedto.ValidateStrIn(&req.Status, true, base.AllAppStatuses, "status")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateAppStatusResp struct {
	Meta *basedto.Meta `json:"meta"`
}
