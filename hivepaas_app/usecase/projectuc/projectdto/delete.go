package projectdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type DeleteProjectReq struct {
	ProjectID string `json:"-"`
}

func NewDeleteProjectReq() *DeleteProjectReq {
	return &DeleteProjectReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *DeleteProjectReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type DeleteProjectResp struct {
	Meta *basedto.Meta `json:"meta"`
}
