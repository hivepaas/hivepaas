package projectdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
)

type DeleteProjectTagsReq struct {
	ProjectID string   `json:"-"`
	Tags      []string `json:"tags"`
}

func NewDeleteProjectTagsReq() *DeleteProjectTagsReq {
	return &DeleteProjectTagsReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *DeleteProjectTagsReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectID")...)
	validators = append(validators, basedto.ValidateSlice(req.Tags, true, 1, nil, "tags")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type DeleteProjectTagsResp struct {
	Meta *basedto.Meta `json:"meta"`
}
