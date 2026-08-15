package periodicjobdto

import (
	"encoding/json"
	"strconv"

	vld "github.com/tiendc/go-validator"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/copier"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/reflectutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type GetPeriodicJobReq struct {
	settings.GetSettingReq
}

func NewGetPeriodicJobReq() *GetPeriodicJobReq {
	return &GetPeriodicJobReq{}
}

func (req *GetPeriodicJobReq) Validate() apperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 5) //nolint:mnd
	validators = append(validators, req.GetSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetPeriodicJobResp struct {
	Meta *basedto.Meta    `json:"meta"`
	Data *PeriodicJobResp `json:"data"`
}

type PeriodicJobResp struct {
	*settings.BaseSettingResp
	Interval     timeutil.Duration         `json:"interval"`
	MaxRetry     int                       `json:"maxRetry"`
	RetryDelay   timeutil.Duration         `json:"retryDelay"`
	Timeout      timeutil.Duration         `json:"timeout"`
	Notification *PeriodicNotificationResp `json:"notification"`
	Healthcheck  *HealthcheckResp          `json:"healthcheck"`
}

type HealthcheckResp struct {
	HealthcheckType base.HealthcheckType `json:"healthcheckType"`
	REST            *HealthcheckRESTResp `json:"rest"`
	GRPC            *HealthcheckGRPCResp `json:"grpc"`
}

type HealthcheckRESTResp struct {
	URL         string                         `json:"url"`
	Method      string                         `json:"method"`
	ContentType string                         `json:"contentType"`
	Body        string                         `json:"body"`
	ReturnCode  string                         `json:"returnCode" copy:"-"`
	ReturnText  *HealthcheckRESTReturnTextResp `json:"returnText"`
	ReturnJSON  *HealthcheckRESTReturnJSONResp `json:"returnJSON"`
}

type HealthcheckRESTReturnTextResp struct {
	Exact string `json:"exact"`
	Regex string `json:"regex"`
}

type HealthcheckRESTReturnJSONResp struct {
	Exact   string `json:"exact" copy:"-"`
	Contain string `json:"contain" copy:"-"`
}

type HealthcheckGRPCResp struct {
	Version      base.HealthcheckGRPCVersion `json:"version"`
	Addr         string                      `json:"addr"`
	Service      string                      `json:"service"`
	ReturnStatus base.HealthcheckGRPCStatus  `json:"returnStatus"`
}

type PeriodicNotificationResp struct {
	*basedto.BaseEventNotificationResp
	MinSendInterval timeutil.Duration `json:"minSendInterval"`
}

func TransformPeriodicJob(
	setting *entity.Setting,
	refObjects *entity.RefObjects,
) (resp *PeriodicJobResp, err error) {
	config := setting.MustAsPeriodicJob()
	if err = copier.Copy(&resp, config); err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp.BaseSettingResp, err = settings.TransformSettingBase(setting)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	TransformHealthcheck(config.Healthcheck, resp.Healthcheck)

	resp.Notification = TransformPeriodicNotification(config.Notification, refObjects)
	return resp, nil
}

func TransformHealthcheck(
	config *entity.PeriodicHealthcheck,
	resp *HealthcheckResp,
) {
	if config == nil || resp == nil {
		return
	}
	TransformHealthcheckREST(config.REST, resp.REST)
}

func TransformHealthcheckREST(
	config *entity.HealthcheckREST,
	resp *HealthcheckRESTResp,
) {
	if config == nil || resp == nil {
		return
	}

	if len(config.ReturnCode) > 0 {
		resp.ReturnCode = gofn.StringJoinBy(config.ReturnCode, ", ", strconv.Itoa)
	}

	if config.ReturnJSON != nil {
		if resp.ReturnJSON == nil {
			resp.ReturnJSON = &HealthcheckRESTReturnJSONResp{}
		}
		if config.ReturnJSON.Exact != nil {
			exact := gofn.Must(json.MarshalIndent(config.ReturnJSON.Exact, "", "   "))
			resp.ReturnJSON.Exact = reflectutil.UnsafeBytesToStr(exact)
		}
		if config.ReturnJSON.Contain != nil {
			contain := gofn.Must(json.MarshalIndent(config.ReturnJSON.Contain, "", "   "))
			resp.ReturnJSON.Contain = reflectutil.UnsafeBytesToStr(contain)
		}
	}
}

func TransformPeriodicNotification(
	config *entity.PeriodicNotification,
	refObjects *entity.RefObjects,
) *PeriodicNotificationResp {
	if config == nil {
		return nil
	}
	return &PeriodicNotificationResp{
		BaseEventNotificationResp: basedto.TransformBaseEventNotification(config.BaseEventNotification, refObjects),
		MinSendInterval:           config.MinSendInterval,
	}
}
