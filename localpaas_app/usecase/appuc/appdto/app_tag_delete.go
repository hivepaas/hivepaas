package appdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
)

type DeleteAppTagsReq struct {
	ProjectID string   `json:"-"`
	AppID     string   `json:"-"`
	Tags      []string `json:"tags"`
}

func NewDeleteAppTagsReq() *DeleteAppTagsReq {
	return &DeleteAppTagsReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *DeleteAppTagsReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectID")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appID")...)
	validators = append(validators, basedto.ValidateSlice(req.Tags, true, 1, nil, "tags")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type DeleteAppTagsResp struct {
	Meta *basedto.Meta `json:"meta"`
}
