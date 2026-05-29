package healthcheckdto

import (
	"encoding/json"
	"math"
	"regexp"
	"strconv"
	"strings"

	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/reflectutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
)

const (
	urlMaxLen = 512

	grpcAddrMaxLen    = 255
	grpcServiceMaxLen = 255
)

type CreateHealthcheckReq struct {
	settings.CreateSettingReq
	*HealthcheckBaseReq
}

type HealthcheckBaseReq struct {
	Name            string                            `json:"name"`
	HealthcheckType base.HealthcheckType              `json:"healthcheckType"`
	Interval        timeutil.Duration                 `json:"interval"`
	MaxRetry        int                               `json:"maxRetry"`
	RetryDelay      timeutil.Duration                 `json:"retryDelay"`
	Timeout         timeutil.Duration                 `json:"timeout"`
	SaveResultTasks bool                              `json:"saveResultTasks"`
	REST            *HealthcheckRESTReq               `json:"rest"`
	GRPC            *HealthcheckGRPCReq               `json:"grpc"`
	Notification    *basedto.BaseEventNotificationReq `json:"notification"`
}

func (req *HealthcheckBaseReq) ToEntity() *entity.Healthcheck {
	res := &entity.Healthcheck{
		HealthcheckType: req.HealthcheckType,
		Interval:        req.Interval,
		MaxRetry:        req.MaxRetry,
		RetryDelay:      req.RetryDelay,
		Timeout:         req.Timeout,
		SaveResultTasks: req.SaveResultTasks,
		Notification:    req.Notification.ToEntity(),
	}
	switch req.HealthcheckType {
	case base.HealthcheckTypeREST:
		res.REST = req.REST.ToEntity()
	case base.HealthcheckTypeGRPC:
		res.GRPC = req.GRPC.ToEntity()
	}
	return res
}

type HealthcheckRESTReq struct {
	URL         string                        `json:"url"`
	Method      base.HTTPMethod               `json:"method"`
	ContentType string                        `json:"contentType"`
	Body        string                        `json:"body"`
	ReturnCode  string                        `json:"returnCode"`
	ReturnText  *HealthcheckRESTReturnTextReq `json:"returnText"`
	ReturnJSON  *HealthcheckRESTReturnJSONReq `json:"returnJSON"`

	// Internal fields
	returnCode []int
}

type HealthcheckRESTReturnTextReq struct {
	Exact string `json:"exact"`
	Regex string `json:"regex"`
}

func (req *HealthcheckRESTReturnTextReq) ToEntity() *entity.HealthcheckRESTReturnText {
	if req == nil {
		return nil
	}
	return &entity.HealthcheckRESTReturnText{
		Exact: req.Exact,
		Regex: req.Regex,
	}
}

type HealthcheckRESTReturnJSONReq struct {
	Exact   string `json:"exact"`
	Contain string `json:"contain"`

	// Internal fields
	exactJSON   any
	containJSON any
}

func (req *HealthcheckRESTReturnJSONReq) ToEntity() *entity.HealthcheckRESTReturnJSON {
	if req == nil {
		return nil
	}
	return &entity.HealthcheckRESTReturnJSON{
		Exact:   req.exactJSON,
		Contain: req.containJSON,
	}
}

func (req *HealthcheckRESTReq) ToEntity() *entity.HealthcheckREST {
	if req == nil {
		return nil
	}
	return &entity.HealthcheckREST{
		URL:         req.URL,
		Method:      req.Method,
		ContentType: req.ContentType,
		Body:        req.Body,
		ReturnCode:  req.returnCode,
		ReturnText:  req.ReturnText.ToEntity(),
		ReturnJSON:  req.ReturnJSON.ToEntity(),
	}
}

func (req *HealthcheckRESTReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStr(&req.URL, true, 1, urlMaxLen, field+"url")...)
	res = append(res, basedto.ValidateStrIn(&req.Method, true, base.AllHTTPMethods, field+"method")...)
	res = append(res, basedto.ValidateValue(len(req.ReturnCode) > 0 || req.ReturnText != nil ||
		req.ReturnJSON != nil, field+"returnCode|returnText|returnJSON")...)

	if len(req.ReturnCode) > 0 {
		items := strings.Split(req.ReturnCode, ",")
		for _, item := range items {
			code, err := strconv.Atoi(strings.TrimSpace(item))
			if err != nil || code < 1 || code > math.MaxInt32 {
				res = append(res, basedto.ValidateValue(false, field+"returnCode")...)
				break
			} else {
				req.returnCode = append(req.returnCode, code)
			}
		}
	}

	if req.ReturnText != nil && req.ReturnText.Regex != "" {
		_, err := regexp.Compile(req.ReturnText.Regex)
		res = append(res, basedto.ValidateValue(err == nil, field+"returnText.regex")...)
	}

	if req.ReturnJSON != nil {
		if req.ReturnJSON.Exact != "" {
			err := json.Unmarshal(reflectutil.UnsafeStrToBytes(req.ReturnJSON.Exact), &req.ReturnJSON.exactJSON)
			res = append(res, basedto.ValidateValue(err == nil, field+"returnJSON.exact")...)
		}
		if req.ReturnJSON.Contain != "" {
			err := json.Unmarshal(reflectutil.UnsafeStrToBytes(req.ReturnJSON.Contain), &req.ReturnJSON.containJSON)
			res = append(res, basedto.ValidateValue(err == nil, field+"returnJSON.contain")...)
		}
	}

	return res
}

type HealthcheckGRPCReq struct {
	Version      base.HealthcheckGRPCVersion `json:"version"`
	Addr         string                      `json:"addr"`
	Service      string                      `json:"service"`
	ReturnStatus base.HealthcheckGRPCStatus  `json:"returnStatus"`
}

func (req *HealthcheckGRPCReq) ToEntity() *entity.HealthcheckGRPC {
	if req == nil {
		return nil
	}
	return &entity.HealthcheckGRPC{
		Version:      req.Version,
		Addr:         req.Addr,
		Service:      req.Service,
		ReturnStatus: req.ReturnStatus,
	}
}

func (req *HealthcheckGRPCReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStrIn(&req.Version, true, base.AllHealthcheckGRPCVersions, field+"version")...)
	res = append(res, basedto.ValidateStr(&req.Addr, true, 1, grpcAddrMaxLen, field+"addr")...)
	res = append(res, basedto.ValidateStr(&req.Service, false, 1, grpcServiceMaxLen, field+"service")...)
	res = append(res, basedto.ValidateNumber(&req.ReturnStatus, true, 1, math.MaxInt32, field+"returnStatus")...)
	return res
}

func (req *HealthcheckBaseReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	switch req.HealthcheckType {
	case base.HealthcheckTypeREST:
		res = append(res, basedto.ValidateValue(req.REST != nil, field+"rest")...)
		res = append(res, req.REST.validate(field+"rest")...)
	case base.HealthcheckTypeGRPC:
		res = append(res, basedto.ValidateValue(req.GRPC != nil, field+"grpc")...)
		res = append(res, req.GRPC.validate(field+"grpc")...)
	}
	return res
}

func NewCreateHealthcheckReq() *CreateHealthcheckReq {
	return &CreateHealthcheckReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *CreateHealthcheckReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type CreateHealthcheckResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *basedto.ObjectIDResp `json:"data"`
}
