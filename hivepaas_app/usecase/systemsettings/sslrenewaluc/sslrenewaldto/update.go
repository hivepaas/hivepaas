package sslrenewaldto

import (
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type UpdateSSLRenewalReq struct {
	settings.UpdateUniqueSettingReq
	*SSLRenewalBaseReq
}

type SSLRenewalBaseReq struct {
	Status       base.SettingStatus                `json:"status"`
	Schedule     ScheduleReq                       `json:"schedule"`
	Notification *basedto.BaseEventNotificationReq `json:"notification"`
}

func (req *SSLRenewalBaseReq) ToEntity() *entity.SSLRenewal {
	return &entity.SSLRenewal{
		Schedule:     req.Schedule.ToEntity(),
		Notification: req.Notification.ToEntity(),
	}
}

func (req *SSLRenewalBaseReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	sched := req.Schedule.ToEntity()
	res = append(res, vld.Must((&sched).IsValid() == nil).OnError(
		vld.SetField(field+"schedule.Interval|schedule.CronExpr", nil),
		vld.SetCustomKey("ERR_VLD_VALUE_REQUIRED_ONLY"),
	))
	res = append(res, basedto.ValidateTime(&req.Schedule.InitialTime, true,
		time.Now().Add(-timeutil.Dur365Days), time.Time{}, field+"schedule.initialTime")...)
	res = append(res, req.Notification.Validate(field+"notification")...)
	return res
}

type ScheduleReq struct {
	CronExpr    string            `json:"cronExpr"` // cronExpr and interval are mutually exclusive
	Interval    timeutil.Duration `json:"interval"`
	InitialTime time.Time         `json:"initialTime"`
}

func (req *ScheduleReq) ToEntity() entity.SchedJobSchedule {
	return entity.SchedJobSchedule{
		CronExpr:    req.CronExpr,
		Interval:    req.Interval,
		InitialTime: req.InitialTime,
	}
}

func NewUpdateSSLRenewalReq() *UpdateSSLRenewalReq {
	return &UpdateSSLRenewalReq{}
}

func (req *UpdateSSLRenewalReq) ModifyRequest() error {
	if req.Schedule.InitialTime.IsZero() {
		req.Schedule.InitialTime = timeutil.NowUTC()
	}
	req.Schedule.InitialTime = req.Schedule.InitialTime.Truncate(time.Minute)
	return nil
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateSSLRenewalReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateSSLRenewalResp struct {
	Meta *basedto.Meta `json:"meta"`
}
