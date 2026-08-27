package projectsettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type DeleteProjectTagsReq struct {
	ProjectID string   `json:"-"`
	Tags      []string `json:"tags"`
}

func NewDeleteProjectTagsReq() *DeleteProjectTagsReq {
	return &DeleteProjectTagsReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *DeleteProjectTagsReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateSlice(req.Tags, true, 1, nil, "tags")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type DeleteProjectTagsResp struct {
	Meta *basedto.Meta `json:"meta"`
}
