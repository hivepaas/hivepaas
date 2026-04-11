package projectsettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
)

type CreateProjectReq struct {
	*ProjectBaseReq
}

type ProjectBaseReq struct {
	Name   string              `json:"name"`
	Status base.ProjectStatus  `json:"status"`
	Tags   []string            `json:"tags"`
	Note   string              `json:"note"`
	Owner  basedto.ObjectIDReq `json:"owner"`
}

func (req *ProjectBaseReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	res = append(res, validateProjectName(&req.Name, field+"name")...)
	res = append(res, basedto.ValidateStrIn(&req.Status, true, base.AllProjectStatuses, field+"status")...)
	res = append(res, validateProjectNote(&req.Note, field+"note")...)
	res = append(res, validateProjectTags(req.Tags, field+"tags")...)
	res = append(res, validateProjectOwner(&req.Owner, field+"owner")...)
	return res
}

func NewCreateProjectReq() *CreateProjectReq {
	return &CreateProjectReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *CreateProjectReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type CreateProjectResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *basedto.ObjectIDResp `json:"data"`
}
