package appsettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
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
func (req *DeleteAppTagsReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	validators = append(validators, basedto.ValidateSlice(req.Tags, true, 1, nil, "tags")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type DeleteAppTagsResp struct {
	Meta *basedto.Meta `json:"meta"`
}
