package appdto

import (
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// GetAppPerformanceMetricsReq asks for an app's routes, or its calls, over a
// range ending now.
type GetAppPerformanceMetricsReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`
	AppID        string `json:"-"`
	Range        string `json:"-" mapstructure:"range"`
}

func NewGetAppPerformanceMetricsReq() *GetAppPerformanceMetricsReq {
	return &GetAppPerformanceMetricsReq{Range: DefaultFunctionMetricsRange}
}

func (req *GetAppPerformanceMetricsReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 5) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	validators = append(validators, basedto.ValidateStrIn(&req.Range, true, FunctionMetricsRanges, "range")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

// AppPerformanceMetricsHeadResp is what an app's routes and its calls both
// say first: whether they can be shown, the range, and how much of the app
// they cover. When Available is false, Reason says why and nothing else is
// set but PreflightReasons.
type AppPerformanceMetricsHeadResp struct {
	Available bool   `json:"available"`
	Reason    string `json:"reason,omitempty"`
	// PreflightReasons say why the nodes the app runs on cannot run OBI, when
	// that is the reason: kernel-too-old, no-btf, container-virt, no-tracefs,
	// lockdown, low-memory.
	PreflightReasons []string `json:"preflightReasons,omitempty"`
	Range            string   `json:"range"`
	// Start and End bound the range: whole steps, End past now.
	Start       time.Time `json:"start"`
	End         time.Time `json:"end"`
	StepSeconds int       `json:"stepSeconds"`
	// Clamped says the range reaches past what the logs keep: it starts where
	// they do.
	Clamped bool `json:"clamped"`
	// Nodes are the nodes the app runs on now, NodesCovered those of them
	// that run OBI: what is served or called on the others is not counted.
	Nodes        int `json:"nodes"`
	NodesCovered int `json:"nodesCovered"`
}

// AppPerformanceCountsResp are requests, or calls, as OBI saw them in the
// app's containers: how many, how many failed, and how long they took in
// milliseconds - p50, p95 and p99, read from buckets summed across replicas,
// close rather than exact, none without one. Past 10 seconds a quantile is
// 10000.
type AppPerformanceCountsResp struct {
	Requests int64    `json:"requests"`
	Errors   int64    `json:"errors"`
	P50      *float64 `json:"p50"`
	P95      *float64 `json:"p95"`
	P99      *float64 `json:"p99"`
}

// AppPerformancePointResp is one step's requests or calls, Time its start,
// and the replicas at its end for an app that autoscales.
type AppPerformancePointResp struct {
	Time time.Time `json:"time"`
	AppPerformanceCountsResp
	Replicas *int `json:"replicas,omitempty"`
}

type GetAppRouteMetricsResp struct {
	Meta *basedto.Meta            `json:"meta"`
	Data *AppRouteMetricsDataResp `json:"data"`
}

// AppRouteMetricsDataResp is what an app served over a range, as OBI saw it
// in its containers: every request, from the proxy or from inside the
// project, by route as the app's framework names it.
type AppRouteMetricsDataResp struct {
	AppPerformanceMetricsHeadResp
	Totals *AppPerformanceCountsResp `json:"totals,omitempty"`
	// Series has one point per step, oldest first.
	Series []*AppPerformancePointResp `json:"series,omitempty"`
	// Routes are the busiest first.
	Routes []*AppRouteResp `json:"routes,omitempty"`
}

// AppRouteResp is the requests to one route: its kind (http or rpc), its
// method, and its route - a template such as /users/{id} when the app's
// framework names one, the path otherwise.
type AppRouteResp struct {
	Kind   string `json:"kind"`
	Method string `json:"method"`
	Route  string `json:"route"`
	AppPerformanceCountsResp
}

type GetAppDependencyMetricsResp struct {
	Meta *basedto.Meta                 `json:"meta"`
	Data *AppDependencyMetricsDataResp `json:"data"`
}

// AppDependencyMetricsDataResp is what an app called over a range, as OBI saw
// it in its containers: other apps, databases, outside hosts.
type AppDependencyMetricsDataResp struct {
	AppPerformanceMetricsHeadResp
	// Kinds are the calls by kind - http, db or rpc - the busiest first.
	Kinds []*AppDependencyKindResp `json:"kinds,omitempty"`
	// Peers are the busiest first.
	Peers []*AppDependencyPeerResp `json:"peers,omitempty"`
}

// AppDependencyKindResp is the calls of one kind: totals, and one point per
// step, oldest first.
type AppDependencyKindResp struct {
	Kind   string                     `json:"kind"`
	Totals *AppPerformanceCountsResp  `json:"totals"`
	Series []*AppPerformancePointResp `json:"series"`
}

// AppDependencyPeerResp is the calls to one peer, and its operations.
type AppDependencyPeerResp struct {
	Kind string `json:"kind"`
	// Peer is what the app called: a host and port for HTTP and RPC, as the
	// app named it; for a database, its system and database, as
	// postgresql/shop.
	Peer string `json:"peer"`
	// App is the env's app behind the peer, when one is: by the name the app
	// called, or a task's or service's address today; for a database, the
	// only one of its engine and database.
	App *AppDependencyAppResp `json:"app,omitempty"`
	AppPerformanceCountsResp
	// Operations are the calls by method - for HTTP and RPC - or operation -
	// for a database, as SELECT - the busiest first.
	Operations []*AppDependencyOperationResp `json:"operations,omitempty"`
}

// AppDependencyAppResp is an app a peer is.
type AppDependencyAppResp struct {
	ID   string `json:"id"`
	Key  string `json:"key"`
	Name string `json:"name"`
}

// AppDependencyOperationResp is the calls to a peer of one method or
// operation.
type AppDependencyOperationResp struct {
	Method    string `json:"method,omitempty"`
	Operation string `json:"operation,omitempty"`
	AppPerformanceCountsResp
}
