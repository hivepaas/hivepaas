package appdto

import (
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// DefaultFunctionMetricsRange is a day: enough to see a function's rhythm - its
// busy hours, its scheduled calls - and what failed overnight.
const DefaultFunctionMetricsRange = "24h"

// FunctionMetricsRanges are the ranges metrics are asked for.
var FunctionMetricsRanges = []string{"1h", "6h", "24h", "7d"}

// GetFunctionMetricsReq asks for a function's calls over a range ending now.
type GetFunctionMetricsReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`
	AppID        string `json:"-"`
	Range        string `json:"-" mapstructure:"range"`
}

func NewGetFunctionMetricsReq() *GetFunctionMetricsReq {
	return &GetFunctionMetricsReq{Range: DefaultFunctionMetricsRange}
}

func (req *GetFunctionMetricsReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 5) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	validators = append(validators, basedto.ValidateStrIn(&req.Range, true, FunctionMetricsRanges, "range")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetFunctionMetricsResp struct {
	Meta *basedto.Meta            `json:"meta"`
	Data *FunctionMetricsDataResp `json:"data"`
}

// FunctionMetricsDataResp is a function's calls over a range. When Available
// is false, Reason says why and nothing else is set.
type FunctionMetricsDataResp struct {
	Available bool   `json:"available"`
	Reason    string `json:"reason,omitempty"`
	Range     string `json:"range"`
	// Start and End bound the range: whole steps, End past now.
	Start       time.Time `json:"start"`
	End         time.Time `json:"end"`
	StepSeconds int       `json:"stepSeconds"`
	// Clamped says the range reaches past what the logs keep: it starts where
	// they do.
	Clamped   bool                       `json:"clamped"`
	Totals    *FunctionMetricsCountsResp `json:"totals,omitempty"`
	ByOutcome map[string]int64           `json:"byOutcome,omitempty"`
	// Series has one point per step, oldest first.
	Series []*FunctionMetricsPointResp `json:"series,omitempty"`
}

// FunctionMetricsCountsResp are a set of calls: how many, how many failed - an
// outcome other than ok - how many the handler answered 5xx, and how long the
// handler ran, in milliseconds, none without a call.
type FunctionMetricsCountsResp struct {
	Calls     int64    `json:"calls"`
	Failed    int64    `json:"failed"`
	Errors5xx int64    `json:"errors5xx"`
	P50       *float64 `json:"p50"`
	P95       *float64 `json:"p95"`
	P99       *float64 `json:"p99"`
}

// FunctionMetricsPointResp is the calls of one step, Time its start.
type FunctionMetricsPointResp struct {
	Time time.Time `json:"time"`
	FunctionMetricsCountsResp
}
