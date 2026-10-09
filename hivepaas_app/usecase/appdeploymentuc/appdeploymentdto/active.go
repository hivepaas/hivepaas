package appdeploymentdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type GetActiveDeploymentReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`
	AppID        string `json:"-"`
}

func NewGetActiveDeploymentReq() *GetActiveDeploymentReq {
	return &GetActiveDeploymentReq{}
}

func (req *GetActiveDeploymentReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetActiveDeploymentResp struct {
	Meta *basedto.Meta `json:"meta"`
	// Data is null when no deployment of the app is queued or running.
	Data *ActiveDeploymentResp `json:"data"`
}

// ActiveDeploymentResp is the app's deployment that has not ended: the one
// running, in-progress, or else the next to run, not-started.
type ActiveDeploymentResp struct {
	ID     string                `json:"id"`
	Status base.DeploymentStatus `json:"status"`
}
