package schedjobdto

import (
	"strings"
	"time"

	vld "github.com/tiendc/go-validator"
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
)

const (
	maxRetryCount = 100
	maxRetryDelay = timeutil.Duration(time.Hour * 24)
	maxTimeout    = timeutil.Duration(time.Hour * 24)
)

type CreateSchedJobReq struct {
	settings.CreateSettingReq
	*SchedJobBaseReq
}

type SchedJobBaseReq struct {
	Name            string                            `json:"name"`
	JobType         base.SchedJobType                 `json:"jobType"`
	Schedule        *ScheduleReq                      `json:"schedule"`
	App             basedto.ObjectIDReq               `json:"app"`
	Priority        base.TaskPriority                 `json:"priority"`
	MaxRetry        int                               `json:"maxRetry"`
	RetryDelay      timeutil.Duration                 `json:"retryDelay"`
	Timeout         timeutil.Duration                 `json:"timeout"`
	ControlDisabled bool                              `json:"controlDisabled"`
	Command         *ContainerCommandReq              `json:"command"`
	Notification    *basedto.BaseEventNotificationReq `json:"notification"`
}

func (req *SchedJobBaseReq) ToEntity() *entity.SchedJob {
	res := &entity.SchedJob{
		JobType:         req.JobType,
		Schedule:        req.Schedule.ToEntity(),
		App:             entity.ObjectID{ID: req.App.ID},
		Priority:        req.Priority,
		MaxRetry:        req.MaxRetry,
		RetryDelay:      req.RetryDelay,
		Timeout:         req.Timeout,
		ControlDisabled: req.ControlDisabled,
		Notification:    req.Notification.ToEntity(),
	}
	if req.JobType == base.SchedJobTypeContainerCommand {
		res.Command = req.Command.ToEntity()
	}
	return res
}

type ScheduleReq struct {
	CronExpr    string            `json:"cronExpr"` // cronExpr and interval are mutually exclusive
	Interval    timeutil.Duration `json:"interval"`
	InitialTime time.Time         `json:"initialTime"`
	EndTime     time.Time         `json:"endTime"`
}

func (req *ScheduleReq) ToEntity() *entity.SchedJobSchedule {
	if req == nil {
		return nil
	}
	return &entity.SchedJobSchedule{
		CronExpr:    req.CronExpr,
		Interval:    req.Interval,
		InitialTime: req.InitialTime,
		EndTime:     req.EndTime,
	}
}

func (req *ScheduleReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateCond(req.ToEntity().IsValid() == nil, field+"cronExpr|interval")...)
	res = append(res, basedto.ValidateTime(&req.InitialTime, true,
		timeutil.NowUTC().Add(-timeutil.Dur365Days), time.Time{}, field+"initialTime")...)
	return res
}

type ContainerCommandReq struct {
	Command     string                 `json:"command"`
	Script      string                 `json:"script"`
	WorkingDir  string                 `json:"workingDir"`
	EnvVars     []*basedto.EnvVarReq   `json:"envVars"`
	ArgGroups   []*CommandArgGroupReq  `json:"argGroups"`
	ConsoleSize *CommandConsoleSizeReq `json:"consoleSize"`
	TTY         bool                   `json:"tty"`
}

func (req *ContainerCommandReq) ToEntity() *entity.SchedJobContainerCommand {
	if req == nil {
		return nil
	}
	return &entity.SchedJobContainerCommand{
		Command:    req.Command,
		Script:     req.Script,
		WorkingDir: req.WorkingDir,
		EnvVars: gofn.MapSlice(req.EnvVars, func(item *basedto.EnvVarReq) *entity.EnvVar {
			return item.ToEntity(false)
		}),
		ArgGroups: gofn.MapSlice(req.ArgGroups, func(item *CommandArgGroupReq) *entity.SchedJobCommandArgGroup {
			return item.ToEntity()
		}),
		ConsoleSize: req.ConsoleSize.ToEntity(),
		TTY:         req.TTY,
	}
}

func (req *ContainerCommandReq) validate(field string) (res []vld.Validator) {
	if req == nil {
		return nil
	}
	if field != "" {
		field += "."
	}
	cmdValid := (req.Command != "" && req.Script == "") || (req.Command == "" && req.Script != "")
	res = append(res, basedto.ValidateCond(cmdValid, field+"command|script")...)
	res = append(res, basedto.ValidateStr(&req.Command, false, 1, int(base.ExecCommandMaxSize), field+"command")...)
	res = append(res, basedto.ValidateStr(&req.Script, false, 1, int(base.ExecCommandMaxSize), field+"script")...)
	// TODO: add validation
	return res
}

type CommandArgGroupReq struct {
	Enabled   bool             `json:"enabled"`
	ExportEnv string           `json:"exportEnv"`
	Separator string           `json:"separator"`
	Args      []*CommandArgReq `json:"args,omitempty"`
}

func (req *CommandArgGroupReq) ToEntity() *entity.SchedJobCommandArgGroup {
	if req == nil {
		return nil
	}
	return &entity.SchedJobCommandArgGroup{
		Enabled:   req.Enabled,
		ExportEnv: req.ExportEnv,
		Separator: req.Separator,
		Args: gofn.MapSlice(req.Args, func(item *CommandArgReq) *entity.SchedJobCommandArg {
			return item.ToEntity()
		}),
	}
}

type CommandArgReq struct {
	Use   bool   `json:"use"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

func (req *CommandArgReq) ToEntity() *entity.SchedJobCommandArg {
	if req == nil {
		return nil
	}
	return &entity.SchedJobCommandArg{
		Use:   req.Use,
		Name:  req.Name,
		Value: req.Value,
	}
}

type CommandConsoleSizeReq struct {
	Width  uint `json:"width"`
	Height uint `json:"height"`
}

func (req *CommandConsoleSizeReq) ToEntity() entity.SchedJobCommandConsoleSize {
	if req == nil {
		return entity.SchedJobCommandConsoleSize{}
	}
	return entity.SchedJobCommandConsoleSize{
		Width:  req.Width,
		Height: req.Height,
	}
}

func (req *SchedJobBaseReq) modifyRequest() error {
	req.Name = strings.TrimSpace(req.Name)
	req.Priority = gofn.Coalesce(req.Priority, base.TaskPriorityDefault)
	if req.Schedule != nil {
		req.Schedule.CronExpr = strings.TrimSpace(req.Schedule.CronExpr)
		if req.Schedule.InitialTime.IsZero() {
			req.Schedule.InitialTime = timeutil.NowUTC()
		}
		req.Schedule.InitialTime = req.Schedule.InitialTime.Truncate(time.Second)
	}
	if req.Command != nil {
		req.Command.Script = strings.ReplaceAll(req.Command.Script, "\r\n", "\n")
	}
	return nil
}

func (req *SchedJobBaseReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStr(&req.Name, true, 1, base.SettingNameMaxLen, field+"name")...)
	res = append(res, basedto.ValidateStrIn(&req.JobType, true, base.AllSchedJobTypes, field+"jobType")...)
	res = append(res, req.Schedule.validate(field+"schedule")...)
	res = append(res, basedto.ValidateObjectIDReq(&req.App, false, field+"app")...)
	res = append(res, basedto.ValidateStrIn(&req.Priority, true, base.AllTaskPriorities, field+"priority")...)
	res = append(res, basedto.ValidateNumber(&req.MaxRetry, false, 1, maxRetryCount, field+"maxRetry")...)
	res = append(res, basedto.ValidateDuration(&req.RetryDelay, false, 1, maxRetryDelay, field+"retryDelay")...)
	res = append(res, basedto.ValidateDuration(&req.Timeout, false, 1, maxTimeout, field+"timeout")...)
	res = append(res, req.Command.validate(field+"command")...)
	res = append(res, req.Notification.Validate(field+"notification")...)
	return res
}

func NewCreateSchedJobReq() *CreateSchedJobReq {
	return &CreateSchedJobReq{}
}

func (req *CreateSchedJobReq) ModifyRequest() error {
	return req.modifyRequest()
}

// Validate implements interface basedto.ReqValidator
func (req *CreateSchedJobReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type CreateSchedJobResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *basedto.ObjectIDResp `json:"data"`
}
