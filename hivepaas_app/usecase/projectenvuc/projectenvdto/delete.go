package projectenvdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
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
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type DeleteProjectEnvResp struct {
	Meta *basedto.Meta `json:"meta"`
}
