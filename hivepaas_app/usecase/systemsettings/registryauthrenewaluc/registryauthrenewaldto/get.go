package registryauthrenewaldto

import (
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

const nextRunsCount = 5

type GetRegistryAuthRenewalReq struct {
	settings.GetUniqueSettingReq
}

func NewGetRegistryAuthRenewalReq() *GetRegistryAuthRenewalReq {
	return &GetRegistryAuthRenewalReq{}
}

func (req *GetRegistryAuthRenewalReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.GetUniqueSettingReq.Validate()...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetRegistryAuthRenewalResp struct {
	Meta *basedto.Meta            `json:"meta"`
	Data *RegistryAuthRenewalResp `json:"data"`
}

type RegistryAuthRenewalResp struct {
	*settings.BaseSettingResp
	Schedule     *ScheduleResp                      `json:"schedule"`
	Notification *basedto.BaseEventNotificationResp `json:"notification"`

	// Calculated fields
	NextRuns []time.Time `json:"nextRuns"`
}

// ScheduleResp is the renewal's schedule: an interval, never a cron.
type ScheduleResp struct {
	Interval    timeutil.Duration `json:"interval"`
	InitialTime time.Time         `json:"initialTime"`
}

func TransformRegistryAuthRenewal(
	setting *entity.Setting,
	refObjects *entity.RefObjects,
) (resp *RegistryAuthRenewalResp, err error) {
	resp = &RegistryAuthRenewalResp{}
	resp.BaseSettingResp, err = settings.TransformSettingBase(setting)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	// An installation that has not made the setting yet answers its default.
	renewal := entity.NewRegistryAuthRenewal(timeutil.NowUTC())
	if setting.Data != "" {
		if renewal, err = setting.AsRegistryAuthRenewal(); err != nil {
			return nil, hperrors.Wrap(err)
		}
	}
	resp.Schedule = &ScheduleResp{
		Interval:    timeutil.Duration(renewal.Interval()),
		InitialTime: renewal.Schedule.InitialTime,
	}
	resp.Notification = basedto.TransformBaseEventNotification(renewal.Notification, refObjects)
	resp.NextRuns, _ = renewal.Schedule.CalcNextRuns(time.Now(), nextRunsCount) //nolint:errcheck

	return resp, nil
}
