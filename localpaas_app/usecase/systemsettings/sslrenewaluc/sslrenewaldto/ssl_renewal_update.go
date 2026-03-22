package sslrenewaldto

import (
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
)

type UpdateSSLRenewalReq struct {
	settings.UpdateSettingReq
	*SSLRenewalBaseReq
}

type SSLRenewalBaseReq struct {
	ScheduleInterval timeutil.Duration `json:"scheduleInterval"`
	ScheduleFrom     time.Time         `json:"scheduleFrom"`
}

func (req *SSLRenewalBaseReq) ToEntity() *entity.SSLRenewal {
	return &entity.SSLRenewal{
		ScheduleInterval: req.ScheduleInterval,
		ScheduleFrom:     req.ScheduleFrom,
	}
}

func (req *SSLRenewalBaseReq) validate(_ string) []vld.Validator {
	// TODO: add validation
	return nil
}

func NewUpdateSSLRenewalReq() *UpdateSSLRenewalReq {
	return &UpdateSSLRenewalReq{}
}

func (req *UpdateSSLRenewalReq) ModifyRequest() error {
	if !req.ScheduleFrom.IsZero() {
		req.ScheduleFrom = req.ScheduleFrom.Truncate(time.Minute)
	}
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
