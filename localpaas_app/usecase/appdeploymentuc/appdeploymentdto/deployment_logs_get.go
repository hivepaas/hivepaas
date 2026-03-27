package appdeploymentdto

import (
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/applog"
)

type GetDeploymentLogsReq struct {
	ProjectID    string        `json:"-"`
	AppID        string        `json:"-"`
	DeploymentID string        `json:"-"`
	Follow       bool          `json:"-" mapstructure:"follow"`
	Since        time.Time     `json:"-" mapstructure:"since"`
	Duration     time.Duration `json:"-" mapstructure:"duration"`
	Tail         int           `json:"-" mapstructure:"tail"`
	Timestamps   bool          `json:"-" mapstructure:"timestamps"`
}

func NewGetDeploymentLogsReq() *GetDeploymentLogsReq {
	return &GetDeploymentLogsReq{}
}

func (req *GetDeploymentLogsReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectID")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appID")...)
	validators = append(validators, basedto.ValidateID(&req.DeploymentID, true, "deploymentID")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetDeploymentLogsResp struct {
	Meta *basedto.Meta           `json:"meta"`
	Data *DeploymentLogsDataResp `json:"data"`
}

type DeploymentLogsDataResp struct {
	Logs          []*applog.LogFrame        `json:"logs"`
	LogChan       <-chan []*applog.LogFrame `json:"-"`
	LogChanCloser func() error              `json:"-"`
}

func TransformDeploymentLogs(logs []*entity.TaskLog) (resp []*applog.LogFrame) {
	resp = make([]*applog.LogFrame, 0, len(logs))
	for _, log := range logs {
		resp = append(resp, &applog.LogFrame{
			Type: log.Type,
			Data: log.Data,
			Ts:   log.Ts,
		})
	}
	return resp
}
