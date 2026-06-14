package taskdto

import (
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/tasklog"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
)

type GetTaskLogsReq struct {
	TaskID     string            `json:"-"`
	Follow     bool              `json:"-" mapstructure:"follow"`
	Since      time.Time         `json:"-" mapstructure:"since"`
	Duration   timeutil.Duration `json:"-" mapstructure:"duration"`
	Tail       int               `json:"-" mapstructure:"tail"`
	Timestamps bool              `json:"-" mapstructure:"timestamps"`
}

func NewGetTaskLogsReq() *GetTaskLogsReq {
	return &GetTaskLogsReq{}
}

func (req *GetTaskLogsReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.TaskID, true, "taskId")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetTaskLogsResp struct {
	Meta *basedto.Meta     `json:"meta"`
	Data *TaskLogsDataResp `json:"data"`
}

type TaskLogsDataResp struct {
	StaticLogs       []*tasklog.LogFrame        `json:"logs"`
	LogsStream       <-chan []*tasklog.LogFrame `json:"-"`
	LogsStreamCloser func() error               `json:"-"`
}

func TransformTaskLogs(logs []*entity.TaskLog) (resp []*tasklog.LogFrame) {
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
