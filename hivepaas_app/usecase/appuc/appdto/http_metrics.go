package appdto

import (
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// GetAppHTTPMetricsReq asks for an app's requests over a range ending now.
type GetAppHTTPMetricsReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`
	AppID        string `json:"-"`
	Range        string `json:"-" mapstructure:"range"`
}

func NewGetAppHTTPMetricsReq() *GetAppHTTPMetricsReq {
	return &GetAppHTTPMetricsReq{Range: DefaultFunctionMetricsRange}
}

func (req *GetAppHTTPMetricsReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 5) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	validators = append(validators, basedto.ValidateStrIn(&req.Range, true, FunctionMetricsRanges, "range")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetAppHTTPMetricsResp struct {
	Meta *basedto.Meta           `json:"meta"`
	Data *AppHTTPMetricsDataResp `json:"data"`
}

// AppHTTPMetricsDataResp is an app's requests over a range, counted from the
// proxy's access log. When Available is false, Reason says why and nothing
// else is set.
type AppHTTPMetricsDataResp struct {
	Available bool   `json:"available"`
	Reason    string `json:"reason,omitempty"`
	Range     string `json:"range"`
	// Start and End bound the range: whole steps, End past now.
	Start       time.Time `json:"start"`
	End         time.Time `json:"end"`
	StepSeconds int       `json:"stepSeconds"`
	// Clamped says the range reaches past what the logs keep: it starts where
	// they do.
	Clamped   bool                  `json:"clamped"`
	Totals    *AppHTTPCountsResp    `json:"totals,omitempty"`
	ByPath    []*AppHTTPPathResp    `json:"byPath,omitempty"`
	ByReplica []*AppHTTPReplicaResp `json:"byReplica,omitempty"`
	// Series has one point per step, oldest first.
	Series []*AppHTTPPointResp `json:"series,omitempty"`
}

// AppHTTPCountsResp are a set of requests: how many, how many the client got
// a 4xx and a 5xx for, how many the proxy could not get to the app at all, and
// how long they took end to end, in milliseconds, none without a request.
type AppHTTPCountsResp struct {
	Requests    int64    `json:"requests"`
	Errors4xx   int64    `json:"errors4xx"`
	Errors5xx   int64    `json:"errors5xx"`
	Unreachable int64    `json:"unreachable"`
	P50         *float64 `json:"p50"`
	P95         *float64 `json:"p95"`
	P99         *float64 `json:"p99"`
}

// AppHTTPPointResp is the requests of one step, Time its start, and the
// replicas at its end for an app that autoscales.
type AppHTTPPointResp struct {
	Time time.Time `json:"time"`
	AppHTTPCountsResp
	Replicas *int `json:"replicas,omitempty"`
}

// AppHTTPPathResp is the requests of one method and path, its numbers and ids
// replaced by :n and :id.
type AppHTTPPathResp struct {
	Method string `json:"method"`
	Path   string `json:"path"`
	AppHTTPCountsResp
}

// AppHTTPReplicaResp is the requests one replica answered, by its address in
// the project's network; an empty address is the requests none answered.
type AppHTTPReplicaResp struct {
	Address string `json:"address"`
	AppHTTPCountsResp
}
