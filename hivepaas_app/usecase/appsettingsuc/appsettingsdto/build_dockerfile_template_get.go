package appsettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
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

func (req *GetDockerfileTemplateReq) Validate() apperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 5) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetDockerfileTemplateResp struct {
	Meta *basedto.Meta           `json:"meta"`
	Data *DockerfileTemplateResp `json:"data"`
}

type DockerfileTemplateResp struct {
	Template string `json:"template"`
}
