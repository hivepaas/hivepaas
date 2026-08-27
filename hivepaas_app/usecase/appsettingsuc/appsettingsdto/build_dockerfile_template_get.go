package appsettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type GetDockerfileTemplateReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`
	AppID        string `json:"-"`
	Type         string `json:"-" mapstructure:"type"`
}

func NewGetDockerfileTemplateReq() *GetDockerfileTemplateReq {
	return &GetDockerfileTemplateReq{}
}

func (req *GetDockerfileTemplateReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetDockerfileTemplateResp struct {
	Meta *basedto.Meta           `json:"meta"`
	Data *DockerfileTemplateResp `json:"data"`
}

type DockerfileTemplateResp struct {
	Template string `json:"template"`
}
