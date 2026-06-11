package sslrenewaldto

import (
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/copier"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
)

type GetSSLRenewalReq struct {
	settings.GetSettingReq
}

func NewGetSSLRenewalReq() *GetSSLRenewalReq {
	return &GetSSLRenewalReq{}
}

func (req *GetSSLRenewalReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.GetSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetSSLRenewalResp struct {
	Meta *basedto.Meta   `json:"meta"`
	Data *SSLRenewalResp `json:"data"`
}

type SSLRenewalResp struct {
	*settings.BaseSettingResp
	Schedule     *ScheduleResp                      `json:"schedule"`
	Notification *basedto.BaseEventNotificationResp `json:"notification"`

	// Calculated fields
	NextRuns []time.Time `json:"nextRuns"`
}

type ScheduleResp struct {
	CronExpr    string            `json:"cronExpr,omitempty"` // cronExpr and interval are mutually exclusive
	Interval    timeutil.Duration `json:"interval,omitempty"`
	InitialTime time.Time         `json:"initialTime"`
}

func TransformSSLRenewal(
	setting *entity.Setting,
	refObjects *entity.RefObjects,
) (resp *SSLRenewalResp, err error) {
	config := setting.MustAsSSLRenewal()
	if err = copier.Copy(&resp, config); err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp.BaseSettingResp, err = settings.TransformSettingBase(setting)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp.Notification = basedto.TransformBaseEventNotification(config.Notification, refObjects)

	// Add next runs
	resp.NextRuns, _ = config.Schedule.CalcNextRuns(time.Now(), 5) //nolint

	return resp, nil
}
