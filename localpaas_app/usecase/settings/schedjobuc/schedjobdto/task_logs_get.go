package schedjobdto

import (
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/tasklog"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
)

type GetSchedJobTaskLogsReq struct {
	settings.BaseSettingReq
	JobID    string            `json:"-"`
	TaskID   string            `json:"-"`
	Follow   bool              `json:"-" mapstructure:"follow"`
	Since    time.Time         `json:"-" mapstructure:"since"`
	Duration timeutil.Duration `json:"-" mapstructure:"duration"`
	Tail     int               `json:"-" mapstructure:"tail"`
}

func NewGetSchedJobTaskLogsReq() *GetSchedJobTaskLogsReq {
	return &GetSchedJobTaskLogsReq{}
}

func (req *GetSchedJobTaskLogsReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.JobID, true, "jobId")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetSchedJobTaskLogsResp struct {
	Meta *basedto.Meta             `json:"meta"`
	Data *SchedJobTaskLogsDataResp `json:"data"`
}

type SchedJobTaskLogsDataResp struct {
	StaticLogs       []*tasklog.LogFrame        `json:"logs"`
	LogsStream       <-chan []*tasklog.LogFrame `json:"-"`
	LogsStreamCloser func() error               `json:"-"`
}

func TransformSchedJobTaskLogs(logs []*entity.TaskLog) (resp []*tasklog.LogFrame) {
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
