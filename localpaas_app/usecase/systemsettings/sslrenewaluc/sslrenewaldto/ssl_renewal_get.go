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
	ScheduleInterval timeutil.Duration `json:"scheduleInterval"`
	ScheduleFrom     time.Time         `json:"scheduleFrom"`
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

	return resp, nil
}
