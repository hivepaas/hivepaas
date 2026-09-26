package appactiondto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type DeployAppReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`
	AppID        string `json:"-"`

	NoCache  bool   `json:"noCache"`
	ChangeID string `json:"changeId"`
}

func NewDeployAppReq() *DeployAppReq {
	return &DeployAppReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *DeployAppReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type DeployAppResp struct {
	Meta *basedto.Meta      `json:"meta"`
	Data *DeployAppDataResp `json:"data"`
}

type DeployAppDataResp struct {
	DeploymentID string `json:"deploymentId"`
	TaskID       string `json:"taskId"`
}
