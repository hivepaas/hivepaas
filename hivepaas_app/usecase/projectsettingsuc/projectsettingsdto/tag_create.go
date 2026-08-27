package projectsettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
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
func (req *CreateProjectTagReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateStr(&req.Tag, true, minTagLen, maxTagLen, "tag")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type CreateProjectTagResp struct {
	Meta *basedto.Meta `json:"meta"`
}
