package appdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
)

type DeleteAppReq struct {
	ProjectID string `json:"-"`
	AppID     string `json:"-"`
}

func NewDeleteAppReq() *DeleteAppReq {
	return &DeleteAppReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *DeleteAppReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectID")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appID")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type DeleteAppResp struct {
	Meta *basedto.Meta `json:"meta"`
}
