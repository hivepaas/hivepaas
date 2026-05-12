package cronjobdto

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

type CreateCronJobReq struct {
	settings.CreateSettingReq
	*CronJobBaseReq
}

type CronJobBaseReq struct {
	Name            string                            `json:"name"`
	CronType        base.CronJobType                  `json:"cronType"`
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

func (req *CronJobBaseReq) ToEntity() *entity.CronJob {
	res := &entity.CronJob{
		CronType:        req.CronType,
		Schedule:        req.Schedule.ToEntity(),
		App:             entity.ObjectID{ID: req.App.ID},
		Priority:        req.Priority,
		MaxRetry:        req.MaxRetry,
		RetryDelay:      req.RetryDelay,
		Timeout:         req.Timeout,
		ControlDisabled: req.ControlDisabled,
		Notification:    req.Notification.ToEntity(),
	}
	if req.CronType == base.CronJobTypeContainerCommand {
		res.Command = req.Command.ToEntity()
	}
	return res
}

type ScheduleReq struct {
	CronExpr    string            `json:"cronExpr"` // cronExpr and interval are mutually exclusive
	Interval    timeutil.Duration `json:"interval"`
	InitialTime time.Time         `json:"initialTime"`
}

func (req *ScheduleReq) ToEntity() *entity.CronJobSchedule {
	if req == nil {
		return nil
	}
	return &entity.CronJobSchedule{
		CronExpr:    req.CronExpr,
		Interval:    req.Interval,
		InitialTime: req.InitialTime,
	}
}

func (req *ScheduleReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateValue(req.ToEntity().IsValid() == nil, field+"cronExpr|interval")...)
	res = append(res, basedto.ValidateTime(&req.InitialTime, true,
		timeutil.NowUTC(), time.Time{}, field+"initialTime")...)
	return res
}

type ContainerCommandReq struct {
	RunInShell string                       `json:"runInShell"`
	Command    string                       `json:"command"`
	WorkingDir string                       `json:"workingDir"`
	EnvVars    []*basedto.EnvVarReq         `json:"envVars"`
	ArgGroups  []*CronJobCommandArgGroupReq `json:"argGroups"`
}

func (req *ContainerCommandReq) ToEntity() *entity.CronJobContainerCommand {
	if req == nil {
		return nil
	}
	return &entity.CronJobContainerCommand{
		Command:    req.Command,
		WorkingDir: req.WorkingDir,
		ArgGroups: gofn.MapSlice(req.ArgGroups, func(item *CronJobCommandArgGroupReq) *entity.CronJobCommandArgGroup {
			return item.ToEntity()
		}),
	}
}

// nolint
func (req *ContainerCommandReq) validate(_ string) (res []vld.Validator) {
	if req == nil {
		return nil
	}
	// TODO: add validation
	return res
}

type CronJobCommandArgGroupReq struct {
	ExportEnv string                  `json:"exportEnv"`
	Separator string                  `json:"separator"`
	Args      []*CronJobCommandArgReq `json:"args,omitempty"`
}

func (req *CronJobCommandArgGroupReq) ToEntity() *entity.CronJobCommandArgGroup {
	if req == nil {
		return nil
	}
	return &entity.CronJobCommandArgGroup{
		ExportEnv: req.ExportEnv,
		Separator: req.Separator,
		Args: gofn.MapSlice(req.Args, func(item *CronJobCommandArgReq) *entity.CronJobCommandArg {
			return item.ToEntity()
		}),
	}
}

type CronJobCommandArgReq struct {
	Use   bool   `json:"use"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

func (req *CronJobCommandArgReq) ToEntity() *entity.CronJobCommandArg {
	if req == nil {
		return nil
	}
	return &entity.CronJobCommandArg{
		Use:   req.Use,
		Name:  req.Name,
		Value: req.Value,
	}
}

func (req *CronJobBaseReq) modifyRequest() error {
	req.Name = strings.TrimSpace(req.Name)
	req.Priority = gofn.Coalesce(req.Priority, base.TaskPriorityDefault)
	if req.Schedule != nil {
		req.Schedule.CronExpr = strings.TrimSpace(req.Schedule.CronExpr)
		if req.Schedule.InitialTime.IsZero() {
			req.Schedule.InitialTime = timeutil.NowUTC().Truncate(time.Second).Add(time.Second)
		}
	}
	return nil
}

func (req *CronJobBaseReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStr(&req.Name, true, 1, base.SettingNameMaxLen, field+"name")...)
	res = append(res, basedto.ValidateStrIn(&req.CronType, true, base.AllCronJobTypes, field+"cronType")...)
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

func NewCreateCronJobReq() *CreateCronJobReq {
	return &CreateCronJobReq{}
}

func (req *CreateCronJobReq) ModifyRequest() error {
	return req.modifyRequest()
}

// Validate implements interface basedto.ReqValidator
func (req *CreateCronJobReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type CreateCronJobResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *basedto.ObjectIDResp `json:"data"`
}
