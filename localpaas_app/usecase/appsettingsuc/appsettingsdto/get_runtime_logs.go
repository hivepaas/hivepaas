package appsettingsdto

import (
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/pkg/applog"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
)

type GetAppRuntimeLogsReq struct {
	ProjectID  string            `json:"-"`
	AppID      string            `json:"-"`
	Follow     bool              `json:"-" mapstructure:"follow"`
	Since      time.Time         `json:"-" mapstructure:"since"`
	Duration   timeutil.Duration `json:"-" mapstructure:"duration"`
	Tail       int               `json:"-" mapstructure:"tail"`
	Timestamps bool              `json:"-" mapstructure:"timestamps"`
}

func NewGetAppRuntimeLogsReq() *GetAppRuntimeLogsReq {
	return &GetAppRuntimeLogsReq{}
}

func (req *GetAppRuntimeLogsReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetAppRuntimeLogsResp struct {
	Meta *basedto.Meta           `json:"meta"`
	Data *AppRuntimeLogsDataResp `json:"data"`
}

type AppRuntimeLogsDataResp struct {
	Logs          []*applog.LogFrame        `json:"logs"`
	LogChan       <-chan []*applog.LogFrame `json:"-"`
	LogChanCloser func() error              `json:"-"`
}
