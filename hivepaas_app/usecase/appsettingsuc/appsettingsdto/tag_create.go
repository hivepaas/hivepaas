package appsettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

const (
	minTagLen = 1
	maxTagLen = 100
)

type CreateAppTagReq struct {
	ProjectID string `json:"-"`
	AppID     string `json:"-"`
	Tag       string `json:"tag"`
}

func NewCreateAppTagReq() *CreateAppTagReq {
	return &CreateAppTagReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *CreateAppTagReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	validators = append(validators, basedto.ValidateStr(&req.Tag, true, minTagLen, maxTagLen, "tag")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type CreateAppTagResp struct {
	Meta *basedto.Meta `json:"meta"`
}
