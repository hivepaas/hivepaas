package projectdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
)

const (
	minTagLen = 1
	maxTagLen = 100
)

type CreateProjectTagReq struct {
	ProjectID string `json:"-"`
	Tag       string `json:"tag"`
}

func NewCreateProjectTagReq() *CreateProjectTagReq {
	return &CreateProjectTagReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *CreateProjectTagReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectID")...)
	validators = append(validators, basedto.ValidateStr(&req.Tag, true, minTagLen, maxTagLen, "tag")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type CreateProjectTagResp struct {
	Meta *basedto.Meta `json:"meta"`
}
