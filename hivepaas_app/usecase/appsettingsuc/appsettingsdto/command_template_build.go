package appsettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type BuildCommandTemplateReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`
	AppID        string `json:"-"`
	CommandID    string `json:"-"`
}

func NewBuildCommandTemplateReq() *BuildCommandTemplateReq {
	return &BuildCommandTemplateReq{}
}

func (req *BuildCommandTemplateReq) Validate() apperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 5) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	validators = append(validators, basedto.ValidateID(&req.CommandID, true, "commandId")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type BuildCommandTemplateResp struct {
	Meta *basedto.Meta     `json:"meta"`
	Data *CommandBuildResp `json:"data"`
}

type CommandBuildResp struct {
	Command string `json:"command"`
}
