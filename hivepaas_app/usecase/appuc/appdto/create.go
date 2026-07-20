package appdto

import (
	"strings"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type CreateAppReq struct {
	ProjectID string `json:"-"`
	*AppBaseReq
}

type AppBaseReq struct {
	Name   string         `json:"name"`
	Status base.AppStatus `json:"status"`
	Env    string         `json:"env"`
	Tags   []string       `json:"tags"`
	Note   string         `json:"note"`
}

func (req *AppBaseReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	res = append(res, validateAppName(&req.Name, field+"name")...)
	res = append(res, basedto.ValidateStrIn(&req.Status, true, base.AllAppStatuses, field+"status")...)
	res = append(res, validateAppEnv(&req.Env, field+"env")...)
	res = append(res, validateAppNote(&req.Note, field+"note")...)
	res = append(res, validateAppTags(req.Tags, field+"tags")...)
	return res
}

func (req *AppBaseReq) modifyRequest() error {
	req.Name = strings.TrimSpace(req.Name)
	req.Env = strings.ToLower(strings.TrimSpace(req.Env))
	return nil
}

func NewCreateAppReq() *CreateAppReq {
	return &CreateAppReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *CreateAppReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, req.validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

func (req *CreateAppReq) ModifyRequest() error {
	return req.modifyRequest()
}

type CreateAppResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *basedto.ObjectIDResp `json:"data"`
}
