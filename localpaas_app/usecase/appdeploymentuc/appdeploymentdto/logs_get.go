package appdeploymentdto

import (
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/tasklog"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
)

type GetDeploymentLogsReq struct {
	Token        string            `json:"-" mapstructure:"token"`
	ProjectID    string            `json:"-"`
	AppID        string            `json:"-"`
	DeploymentID string            `json:"-"`
	Follow       bool              `json:"-" mapstructure:"follow"`
	Since        time.Time         `json:"-" mapstructure:"since"`
	Duration     timeutil.Duration `json:"-" mapstructure:"duration"`
	Tail         int               `json:"-" mapstructure:"tail"`
	Timestamps   bool              `json:"-" mapstructure:"timestamps"`
}

func NewGetDeploymentLogsReq() *GetDeploymentLogsReq {
	return &GetDeploymentLogsReq{}
}

func (req *GetDeploymentLogsReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	validators = append(validators, basedto.ValidateID(&req.DeploymentID, true, "deploymentId")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetDeploymentLogsResp struct {
	Meta *basedto.Meta           `json:"meta"`
	Data *DeploymentLogsDataResp `json:"data"`
}

type DeploymentLogsDataResp struct {
	StaticLogs         []*tasklog.LogFrame        `json:"logs"`
	RealtimeLogsStream <-chan []*tasklog.LogFrame `json:"-"`
	LogsStreamCloser   func() error               `json:"-"`
}

func TransformDeploymentLogs(logs []*entity.TaskLog) (resp []*tasklog.LogFrame) {
	resp = make([]*tasklog.LogFrame, 0, len(logs))
	for _, log := range logs {
		resp = append(resp, &tasklog.LogFrame{
			Type: log.Type,
			Data: log.Data,
			Ts:   log.Ts,
		})
	}
	return resp
}
