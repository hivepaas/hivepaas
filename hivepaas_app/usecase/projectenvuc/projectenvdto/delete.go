package projectenvdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type DeleteProjectEnvReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`
}

func NewDeleteProjectEnvReq() *DeleteProjectEnvReq {
	return &DeleteProjectEnvReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *DeleteProjectEnvReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type DeleteProjectEnvResp struct {
	Meta *basedto.Meta `json:"meta"`
}
