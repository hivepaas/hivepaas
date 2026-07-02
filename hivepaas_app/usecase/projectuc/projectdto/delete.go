package projectdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type DeleteProjectReq struct {
	ProjectID string `json:"-"`
}

func NewDeleteProjectReq() *DeleteProjectReq {
	return &DeleteProjectReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *DeleteProjectReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type DeleteProjectResp struct {
	Meta *basedto.Meta `json:"meta"`
}
