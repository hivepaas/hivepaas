package projectenvdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type DeleteProjectEnvReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`
}

func NewDeleteProjectEnvReq() *DeleteProjectEnvReq {
	return &DeleteProjectEnvReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *DeleteProjectEnvReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateStr(&req.ProjectEnvID, true,
		base.ProjectEnvMinLen, base.ProjectEnvMaxLen, "projectEnv")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type DeleteProjectEnvResp struct {
	Meta *basedto.Meta `json:"meta"`
}
