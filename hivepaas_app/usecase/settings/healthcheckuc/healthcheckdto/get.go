package healthcheckdto

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

type GetHealthcheckReq struct {
	settings.GetSettingReq
}

func NewGetHealthcheckReq() *GetHealthcheckReq {
	return &GetHealthcheckReq{}
}

func (req *GetHealthcheckReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.GetSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetHealthcheckResp struct {
	Meta *basedto.Meta    `json:"meta"`
	Data *HealthcheckResp `json:"data"`
}

type HealthcheckResp struct {
	*settings.BaseSettingResp
	HealthcheckType base.HealthcheckType         `json:"healthcheckType"`
	Interval        timeutil.Duration            `json:"interval"`
	MaxRetry        int                          `json:"maxRetry"`
	RetryDelay      timeutil.Duration            `json:"retryDelay"`
	Timeout         timeutil.Duration            `json:"timeout"`
	SaveResultTasks bool                         `json:"saveResultTasks"`
	REST            *HealthcheckRESTResp         `json:"rest"`
	GRPC            *HealthcheckGRPCResp         `json:"grpc"`
	Notification    *HealthcheckNotificationResp `json:"notification"`
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

type HealthcheckNotificationResp struct {
	*basedto.BaseEventNotificationResp
	MinSendInterval timeutil.Duration `json:"minSendInterval"`
}

func TransformHealthcheck(
	setting *entity.Setting,
	refObjects *entity.RefObjects,
) (resp *HealthcheckResp, err error) {
	config := setting.MustAsHealthcheck()
	if err = copier.Copy(&resp, config); err != nil {
		return nil, apperrors.New(err)
	}

	resp.BaseSettingResp, err = settings.TransformSettingBase(setting)
	if err != nil {
		return nil, apperrors.New(err)
	}

	TransformHealthcheckREST(config.REST, resp.REST)

	resp.Notification = TransformHealthcheckNotification(config.Notification, refObjects)
	return resp, nil
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

func TransformHealthcheckNotification(
	config *entity.HealthcheckNotification,
	refObjects *entity.RefObjects,
) *HealthcheckNotificationResp {
	if config == nil {
		return nil
	}
	return &HealthcheckNotificationResp{
		BaseEventNotificationResp: basedto.TransformBaseEventNotification(config.BaseEventNotification, refObjects),
		MinSendInterval:           config.MinSendInterval,
	}
}
