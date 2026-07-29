package projectenvdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type UpdateProjectEnvStatusReq struct {
	ProjectID    string             `json:"-"`
	ProjectEnvID string             `json:"-"`
	UpdateVer    int                `json:"updateVer"`
	Status       base.ProjectStatus `json:"status"`
}

func NewUpdateProjectEnvStatusReq() *UpdateProjectEnvStatusReq {
	return &UpdateProjectEnvStatusReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateProjectEnvStatusReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateStr(&req.ProjectEnvID, true,
		base.ProjectEnvMinLen, base.ProjectEnvMaxLen, "projectEnv")...)
	validators = append(validators, basedto.ValidateStrIn(&req.Status, true, base.AllProjectStatuses, "status")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateProjectEnvStatusResp struct {
	Meta *basedto.Meta `json:"meta"`
}
