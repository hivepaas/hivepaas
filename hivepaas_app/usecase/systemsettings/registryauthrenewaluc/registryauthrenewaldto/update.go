package registryauthrenewaldto

import (
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type UpdateRegistryAuthRenewalReq struct {
	settings.UpdateUniqueSettingReq
	*RegistryAuthRenewalBaseReq
}

type RegistryAuthRenewalBaseReq struct {
	Status       base.SettingStatus                `json:"status"`
	Schedule     ScheduleReq                       `json:"schedule"`
	Notification *basedto.BaseEventNotificationReq `json:"notification"`
}

// ScheduleReq is an interval from 1 to 10 hours: an ECR token lives 12, and one
// handed over must outlive the next run by an hour.
type ScheduleReq struct {
	Interval    timeutil.Duration `json:"interval"`
	InitialTime time.Time         `json:"initialTime"`
}

func (req *RegistryAuthRenewalBaseReq) ToEntity() *entity.RegistryAuthRenewal {
	return &entity.RegistryAuthRenewal{
		Schedule: entity.SchedJobSchedule{
			Interval:    req.Schedule.Interval,
			InitialTime: req.Schedule.InitialTime,
		},
		Notification: req.Notification.ToEntity(),
	}
}

func (req *RegistryAuthRenewalBaseReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStrIn(&req.Status, true,
		[]base.SettingStatus{base.SettingStatusActive, base.SettingStatusDisabled}, field+"status")...)
	res = append(res, basedto.ValidateDuration(&req.Schedule.Interval, true,
		timeutil.Duration(entity.RegistryAuthRenewalIntervalMin),
		timeutil.Duration(entity.RegistryAuthRenewalIntervalMax), field+"schedule.interval")...)
	res = append(res, basedto.ValidateTime(&req.Schedule.InitialTime, true,
		time.Now().Add(-timeutil.Dur365Days), time.Time{}, field+"schedule.initialTime")...)
	res = append(res, req.Notification.Validate(field+"notification")...)
	return res
}

func NewUpdateRegistryAuthRenewalReq() *UpdateRegistryAuthRenewalReq {
	return &UpdateRegistryAuthRenewalReq{}
}

func (req *UpdateRegistryAuthRenewalReq) ModifyRequest() error {
	if req.Schedule.InitialTime.IsZero() {
		req.Schedule.InitialTime = timeutil.NowUTC()
	}
	req.Schedule.InitialTime = req.Schedule.InitialTime.Truncate(time.Minute)
	return nil
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateRegistryAuthRenewalReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.UpdateUniqueSettingReq.Validate()...)
	validators = append(validators, req.validate("")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateRegistryAuthRenewalResp struct {
	Meta *basedto.Meta `json:"meta"`
}
