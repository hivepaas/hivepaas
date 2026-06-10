package appdto

import (
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/pkg/tasklog"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
)

type GetAppLogsReq struct {
	Token      string            `json:"-" mapstructure:"token"`
	ProjectID  string            `json:"-"`
	AppID      string            `json:"-"`
	TaskID     string            `json:"-" mapstructure:"taskId"`
	Follow     bool              `json:"-" mapstructure:"follow"`
	Since      time.Time         `json:"-" mapstructure:"since"`
	Duration   timeutil.Duration `json:"-" mapstructure:"duration"`
	Tail       int               `json:"-" mapstructure:"tail"`
	Timestamps *bool             `json:"-" mapstructure:"timestamps"`
}

func NewGetAppLogsReq() *GetAppLogsReq {
	return &GetAppLogsReq{}
}

func (req *GetAppLogsReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetAppLogsResp struct {
	Meta *basedto.Meta    `json:"meta"`
	Data *AppLogsDataResp `json:"data"`
}

type AppLogsDataResp struct {
	StaticLogs       []*tasklog.LogFrame        `json:"logs"`
	LogsStream       <-chan []*tasklog.LogFrame `json:"-"`
	LogsStreamCloser func() error               `json:"-"`
}
