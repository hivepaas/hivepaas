package loggingdto

import (
	"fmt"
	"time"

	vld "github.com/tiendc/go-validator"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/obi"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

// performanceNodesMax is a sanity bound on the nodes a request lists: more
// than a swarm runs.
const performanceNodesMax = 1000

// The reasons the nodes' statuses are not read.
const (
	// PerformanceStatusReasonLogsNotStored: OBI runs, and its agents write
	// their statuses, only while the logs are stored.
	PerformanceStatusReasonLogsNotStored = "logs-not-stored"
	// PerformanceStatusReasonUnreadable: the stored logs could not be read.
	PerformanceStatusReasonUnreadable = "unreadable"
)

type GetLoggingPerformanceReq struct{}

func NewGetLoggingPerformanceReq() *GetLoggingPerformanceReq {
	return &GetLoggingPerformanceReq{}
}

func (req *GetLoggingPerformanceReq) Validate() hperrors.ValidationErrors {
	return nil
}

type GetLoggingPerformanceResp struct {
	Meta *basedto.Meta           `json:"meta"`
	Data *LoggingPerformanceResp `json:"data"`
}

// LoggingPerformanceResp is the collection of apps' routes and calls by OBI:
// whether it is on, the swarm's nodes, and the capacities a node can be given.
type LoggingPerformanceResp struct {
	Enabled bool `json:"enabled"`
	// Configured says the logging settings were saved: these are saved with
	// them, and cannot be before.
	Configured bool `json:"configured"`
	// LogsStored says the logs are stored: OBI runs only while they are, its
	// numbers being rows in them.
	LogsStored bool `json:"logsStored"`
	// UpdateVer is the logging settings' version, which an update sends back:
	// saving these changes it, as saving the logging settings does.
	UpdateVer int `json:"updateVer"`
	// StatusReason says why the nodes' statuses were not read:
	// logs-not-stored or unreadable. Empty when they were.
	StatusReason string                     `json:"statusReason,omitempty"`
	Capacities   []*PerformanceCapacityResp `json:"capacities"`
	Nodes        []*PerformanceNodeResp     `json:"nodes"`
}

// PerformanceCapacityResp is a capacity a node can be given: how much memory
// OBI takes at it, and about how many requests and connections it tracks at
// once.
type PerformanceCapacityResp struct {
	Capacity  string `json:"capacity"`
	MemoryMiB int    `json:"memoryMiB"`
	Tracked   int    `json:"tracked"`
}

// PerformanceNodeResp is a node of the swarm: whether it runs OBI and at
// which capacity, and what its agent last said of it.
type PerformanceNodeResp struct {
	ID           string `json:"id"`
	Hostname     string `json:"hostname"`
	Role         string `json:"role"`
	State        string `json:"state"`
	Availability string `json:"availability"`
	MemoryBytes  int64  `json:"memoryBytes"`
	// Recommended is the capacity HivePaaS recommends for the node's memory.
	Recommended string `json:"recommended"`
	Enabled     bool   `json:"enabled"`
	// Capacity is the one chosen for the node: auto for the recommended one.
	Capacity string `json:"capacity"`
	// Status is what the node's agent last said - every minute while the
	// feature is on, every 10 minutes while it is off; nil when it said nothing:
	// the logs not stored, or an agent from before.
	Status *PerformanceNodeStatusResp `json:"status,omitempty"`
}

// PerformanceNodeStatusResp is a node's status as its agent wrote it: whether
// OBI is wanted and runs there, for how many apps, and whether the node can
// run it, with why not.
type PerformanceNodeStatusResp struct {
	Time    time.Time `json:"time"`
	Wanted  bool      `json:"wanted"`
	Running bool      `json:"running"`
	Apps    int       `json:"apps"`
	// OK says the node can run OBI; Reasons why not: kernel-too-old, no-btf,
	// container-virt, no-tracefs, lockdown, low-memory.
	OK             bool     `json:"ok"`
	Reasons        []string `json:"reasons,omitempty"`
	Kernel         string   `json:"kernel,omitempty"`
	MemTotalMB     int      `json:"memTotalMb,omitempty"`
	MemAvailableMB int      `json:"memAvailableMb,omitempty"`
	// Capacity is the one OBI runs, or would run, with there: the one chosen,
	// or the one recommended for the memory the node reads.
	Capacity string `json:"capacity"`
}

// PerformanceCapacities are the capacities a node can be given, smallest
// first.
func PerformanceCapacities() []*PerformanceCapacityResp {
	return gofn.MapSlice(obi.Capacities, func(c obi.Capacity) *PerformanceCapacityResp {
		return &PerformanceCapacityResp{Capacity: string(c), MemoryMiB: c.MemoryMiB(), Tracked: c.Tracked()}
	})
}

// PerformanceNodeStatus is a status as the API answers it.
func PerformanceNodeStatus(at time.Time, s *obi.Status) *PerformanceNodeStatusResp {
	return &PerformanceNodeStatusResp{Time: at, Wanted: s.Wanted, Running: s.Running, Apps: s.Apps,
		OK: s.Preflight.OK, Reasons: s.Preflight.Reasons, Kernel: s.Preflight.Kernel,
		MemTotalMB: s.Preflight.MemTotalMB, MemAvailableMB: s.Preflight.MemAvailableMB,
		Capacity: string(s.Preflight.Capacity)}
}

type UpdateLoggingPerformanceReq struct {
	settings.UpdateUniqueSettingReq
	Enabled bool                  `json:"enabled"`
	Nodes   []*PerformanceNodeReq `json:"nodes"`
}

// PerformanceNodeReq is a node that runs OBI, and its capacity: small, medium
// or large; auto, or none, for the one recommended for its memory.
type PerformanceNodeReq struct {
	ID       string `json:"id"`
	Capacity string `json:"capacity"`
}

func NewUpdateLoggingPerformanceReq() *UpdateLoggingPerformanceReq {
	return &UpdateLoggingPerformanceReq{}
}

func (req *UpdateLoggingPerformanceReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 2+2*len(req.Nodes)) //nolint:mnd // the list, then each node
	ids := make([]string, 0, len(req.Nodes))
	for _, node := range req.Nodes {
		if node == nil {
			continue
		}
		ids = append(ids, node.ID)
	}
	validators = append(validators, basedto.ValidateIDSlice(ids, true, 0, "nodes")...)
	validators = append(validators, basedto.ValidateSliceEx(req.Nodes, false, 0, performanceNodesMax, nil,
		"nodes")...)
	capacities := append([]string{"", string(obi.CapacityAuto)}, gofn.MapSlice(obi.Capacities,
		func(c obi.Capacity) string { return string(c) })...)
	for i, node := range req.Nodes {
		if node == nil {
			continue
		}
		validators = append(validators, basedto.ValidateID(&node.ID, true, fmt.Sprintf("nodes[%d].id", i))...)
		validators = append(validators, basedto.ValidateStrIn(&node.Capacity, false, capacities,
			fmt.Sprintf("nodes[%d].capacity", i))...)
	}
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

// ToEntity is the nodes as stored: each once, its capacity auto when none.
func (req *UpdateLoggingPerformanceReq) ToEntity() *entity.LoggingPerformance {
	out := &entity.LoggingPerformance{Enabled: req.Enabled}
	for _, node := range req.Nodes {
		if node == nil {
			continue
		}
		out.Nodes = append(out.Nodes, &entity.LoggingPerformanceNode{ID: node.ID,
			Capacity: gofn.Coalesce(node.Capacity, string(obi.CapacityAuto))})
	}
	return out
}

type UpdateLoggingPerformanceResp struct {
	Meta *basedto.Meta `json:"meta"`
}
