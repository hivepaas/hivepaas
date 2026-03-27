package appdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
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
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectID")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appID")...)
	validators = append(validators, basedto.ValidateStr(&req.Tag, true, minTagLen, maxTagLen, "tag")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type CreateAppTagResp struct {
	Meta *basedto.Meta `json:"meta"`
}
