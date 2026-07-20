package projectdto

import (
	"fmt"
	"strings"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

const (
	projectNameMinLen = 1
	projectNameMaxLen = 100

	projectEnvMinLen = 1
	projectEnvMaxLen = 50

	projectTagMinLen = 0
	projectTagMaxLen = 50

	projectNoteMinLen = 1
	projectNoteMaxLen = 10000
)

type CreateProjectReq struct {
	*ProjectBaseReq
}

type ProjectBaseReq struct {
	Name   string              `json:"name"`
	Status base.ProjectStatus  `json:"status"`
	Envs   []*ProjectEnvReq    `json:"envs"`
	Tags   []string            `json:"tags"`
	Note   string              `json:"note"`
	Owner  basedto.ObjectIDReq `json:"owner"`
}

func (req *ProjectBaseReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	res = append(res, req.validateName(field+"name")...)
	res = append(res, basedto.ValidateStrIn(&req.Status, true, base.AllProjectStatuses, field+"status")...)
	res = append(res, basedto.ValidateStr(&req.Note, false, projectNoteMinLen, projectNoteMaxLen, field)...)
	res = append(res, req.validateEnvs(field+"envs")...)
	res = append(res, basedto.ValidateSliceEx(req.Tags, true, projectTagMinLen, projectTagMaxLen, nil, field)...)
	res = append(res, basedto.ValidateObjectIDReq(&req.Owner, false, field+"owner")...)
	return res
}

func (req *ProjectBaseReq) validateName(field string) []vld.Validator {
	return basedto.ValidateStr(&req.Name, true, projectNameMinLen, projectNameMaxLen, field)
	// TODO: need validation for valid characters
}

func (req *ProjectBaseReq) validateEnvs(field string) (res []vld.Validator) {
	for i, env := range req.Envs {
		res = append(res, basedto.ValidateStr(&env.Name, true, projectEnvMinLen, projectEnvMaxLen,
			field+fmt.Sprintf("[%v].name", i))...)
	}
	res = append(res, vld.SliceUniqueBy(req.Envs, func(env *ProjectEnvReq) string { return env.Name }))
	return res
}

func (req *ProjectBaseReq) modifyRequest() error {
	for _, env := range req.Envs {
		env.Name = strings.ToLower(strings.TrimSpace(env.Name))
	}
	return nil
}

type ProjectEnvReq struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

func (req *ProjectEnvReq) ToEntity() *entity.Env {
	if req == nil {
		return nil
	}
	return &entity.Env{
		Name:  req.Name,
		Color: req.Color,
	}
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

func (req *CreateProjectReq) ModifyRequest() error {
	return req.modifyRequest()
}

type CreateProjectResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *basedto.ObjectIDResp `json:"data"`
}
