package appdto

import (
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// GetAppResourceMetricsReq asks for an app's containers' usage over a range
// ending now.
type GetAppResourceMetricsReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`
	AppID        string `json:"-"`
	Range        string `json:"-" mapstructure:"range"`
}

func NewGetAppResourceMetricsReq() *GetAppResourceMetricsReq {
	return &GetAppResourceMetricsReq{Range: DefaultFunctionMetricsRange}
}

func (req *GetAppResourceMetricsReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 5) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	validators = append(validators, basedto.ValidateStrIn(&req.Range, true, FunctionMetricsRanges, "range")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetAppResourceMetricsResp struct {
	Meta *basedto.Meta               `json:"meta"`
	Data *AppResourceMetricsDataResp `json:"data"`
}

// AppResourceMetricsDataResp is an app's containers' usage over a range, read
// from the rows the agent writes every 15 seconds. When Available is false,
// Reason says why and nothing else is set.
type AppResourceMetricsDataResp struct {
	Available   bool      `json:"available"`
	Reason      string    `json:"reason,omitempty"`
	Range       string    `json:"range"`
	Start       time.Time `json:"start"`
	End         time.Time `json:"end"`
	StepSeconds int       `json:"stepSeconds"`
	Clamped     bool      `json:"clamped"`
	// Totals are over the range: CPU's average and peak, memory's peak, the
	// limits last seen, the OOM kills.
	Totals *AppResourceTotalsResp `json:"totals,omitempty"`
	// Series has one point per step, oldest first: the containers summed.
	Series []*AppResourcePointResp `json:"series,omitempty"`
	// Containers are the app's containers over the range, the last seen first.
	Containers []*AppResourceContainerResp `json:"containers,omitempty"`
}

// AppResourceTotalsResp are an app's usage over a range: CPU in cores, memory
// in bytes, a limit 0 when there is none. CPU and memory are nil without a row.
type AppResourceTotalsResp struct {
	CPU         *float64 `json:"cpu"`
	CPUPeak     *float64 `json:"cpuPeak"`
	CPULimit    float64  `json:"cpuLimit"`
	MemoryPeak  *float64 `json:"memoryPeak"`
	MemoryLimit float64  `json:"memoryLimit"`
	OOMKills    int64    `json:"oomKills"`
}

// AppResourcePointResp is one step's usage, Time its start: CPU in cores and
// memory in bytes, nil without a row; network and disk in bytes a second.
type AppResourcePointResp struct {
	Time        time.Time `json:"time"`
	CPU         *float64  `json:"cpu"`
	CPULimit    float64   `json:"cpuLimit"`
	Memory      *float64  `json:"memory"`
	MemoryLimit float64   `json:"memoryLimit"`
	OOMKills    int64     `json:"oomKills"`
	NetRx       float64   `json:"netRx"`
	NetTx       float64   `json:"netTx"`
	IORead      float64   `json:"ioRead"`
	IOWrite     float64   `json:"ioWrite"`
}

// AppResourceContainerResp is one container over the range, by its short id.
type AppResourceContainerResp struct {
	Container   string    `json:"container"`
	CPU         float64   `json:"cpu"`
	CPUPeak     float64   `json:"cpuPeak"`
	Memory      float64   `json:"memory"`
	MemoryLimit float64   `json:"memoryLimit"`
	OOMKills    int64     `json:"oomKills"`
	LastSeen    time.Time `json:"lastSeen"`
}
