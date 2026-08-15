package appdto

import (
	"strings"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type CreateAppReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`
	*AppBaseReq
}

type AppBaseReq struct {
	Name   string         `json:"name"`
	Status base.AppStatus `json:"status"`
	Tags   []string       `json:"tags"`
	Note   string         `json:"note"`
}

func (req *AppBaseReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	res = append(res, validateAppName(&req.Name, field+"name")...)
	res = append(res, basedto.ValidateStrIn(&req.Status, true, base.AllAppStatuses, field+"status")...)
	res = append(res, validateAppNote(&req.Note, field+"note")...)
	res = append(res, validateAppTags(req.Tags, field+"tags")...)
	return res
}

func (req *AppBaseReq) modifyRequest() error {
	req.Name = strings.TrimSpace(req.Name)
	return nil
}

func NewCreateAppReq() *CreateAppReq {
	return &CreateAppReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *CreateAppReq) Validate() apperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 5) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
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
