package appdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
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
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type DeleteAppResp struct {
	Meta *basedto.Meta `json:"meta"`
}
