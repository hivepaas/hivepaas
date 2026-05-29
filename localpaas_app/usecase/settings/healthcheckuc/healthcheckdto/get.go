package healthcheckdto

import (
	"encoding/json"
	"strconv"

	vld "github.com/tiendc/go-validator"
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/copier"
	"github.com/localpaas/localpaas/localpaas_app/pkg/reflectutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
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
	HealthcheckType base.HealthcheckType               `json:"healthcheckType"`
	Interval        timeutil.Duration                  `json:"interval"`
	MaxRetry        int                                `json:"maxRetry"`
	RetryDelay      timeutil.Duration                  `json:"retryDelay"`
	Timeout         timeutil.Duration                  `json:"timeout"`
	SaveResultTasks bool                               `json:"saveResultTasks"`
	REST            *HealthcheckRESTResp               `json:"rest"`
	GRPC            *HealthcheckGRPCResp               `json:"grpc"`
	Notification    *basedto.BaseEventNotificationResp `json:"notification"`
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

func TransformHealthcheck(
	setting *entity.Setting,
	refObjects *entity.RefObjects,
) (resp *HealthcheckResp, err error) {
	config := setting.MustAsHealthcheck()
	if err = copier.Copy(&resp, config); err != nil {
		return nil, apperrors.Wrap(err)
	}

	if config.REST != nil { //nolint:nestif
		if len(config.REST.ReturnCode) > 0 {
			resp.REST.ReturnCode = gofn.StringJoinBy(config.REST.ReturnCode, ", ", strconv.Itoa)
		}
		if config.REST.ReturnJSON != nil {
			if resp.REST.ReturnJSON == nil {
				resp.REST.ReturnJSON = &HealthcheckRESTReturnJSONResp{}
			}
			if config.REST.ReturnJSON.Exact != nil {
				exact := gofn.Must(json.Marshal(config.REST.ReturnJSON.Exact))
				resp.REST.ReturnJSON.Exact = reflectutil.UnsafeBytesToStr(exact)
			}
			if config.REST.ReturnJSON.Contain != nil {
				contain := gofn.Must(json.Marshal(config.REST.ReturnJSON.Contain))
				resp.REST.ReturnJSON.Contain = reflectutil.UnsafeBytesToStr(contain)
			}
		}
	}

	resp.BaseSettingResp, err = settings.TransformSettingBase(setting)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp.Notification = basedto.TransformBaseEventNotification(config.Notification, refObjects)
	return resp, nil
}
