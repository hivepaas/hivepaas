package taskdto

import (
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/applog"
)

type GetTaskLogsReq struct {
	TaskID     string        `json:"-"`
	Follow     bool          `json:"-" mapstructure:"follow"`
	Since      time.Time     `json:"-" mapstructure:"since"`
	Duration   time.Duration `json:"-" mapstructure:"duration"`
	Tail       int           `json:"-" mapstructure:"tail"`
	Timestamps bool          `json:"-" mapstructure:"timestamps"`
}

func NewGetTaskLogsReq() *GetTaskLogsReq {
	return &GetTaskLogsReq{}
}

func (req *GetTaskLogsReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.TaskID, true, "taskID")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetTaskLogsResp struct {
	Meta *basedto.Meta     `json:"meta"`
	Data *TaskLogsDataResp `json:"data"`
}

type TaskLogsDataResp struct {
	Logs          []*applog.LogFrame        `json:"logs"`
	LogChan       <-chan []*applog.LogFrame `json:"-"`
	LogChanCloser func() error              `json:"-"`
}

func TransformTaskLogs(logs []*entity.TaskLog) (resp []*applog.LogFrame) {
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
