package appsettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type DeleteAppTagsReq struct {
	ProjectID    string   `json:"-"`
	ProjectEnvID string   `json:"-"`
	AppID        string   `json:"-"`
	Tags         []string `json:"tags"`
}

func NewDeleteAppTagsReq() *DeleteAppTagsReq {
	return &DeleteAppTagsReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *DeleteAppTagsReq) Validate() apperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 5) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	validators = append(validators, basedto.ValidateSlice(req.Tags, true, 1, nil, "tags")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type DeleteAppTagsResp struct {
	Meta *basedto.Meta `json:"meta"`
}
