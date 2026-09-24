package homedto

import (
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/attentionservice"
)

type GetHomeAttentionResp struct {
	Meta *basedto.Meta          `json:"meta"`
	Data *HomeAttentionDataResp `json:"data"`
}

type HomeAttentionDataResp struct {
	// Items is what needs attention, the most severe first.
	Items []*AttentionItemResp `json:"items"`
}

// AttentionItemResp is one thing that needs attention. The fields after Subject
// are filled by the kinds they belong to, for the screen to word.
type AttentionItemResp struct {
	// Kind is app-not-running, app-restarting, node-down or node-overcommitted.
	Kind string `json:"kind"`
	// Severity is critical or warning.
	Severity string `json:"severity"`
	// Scope is where it is fixed: app (the app's screens), system or cluster.
	Scope   string                   `json:"scope"`
	Project *basedto.NamedObjectResp `json:"project,omitempty"`
	Env     string                   `json:"env,omitempty"`
	App     *basedto.NamedObjectResp `json:"app,omitempty"`
	// Subject names what it is about: an app, a system service, a node.
	Subject string `json:"subject"`

	Running           uint64 `json:"running,omitempty"`
	Desired           uint64 `json:"desired,omitempty"`
	Restarts          int    `json:"restarts,omitempty"`
	LastError         string `json:"lastError,omitempty"`
	NodeState         string `json:"nodeState,omitempty"`
	MemoryLimitsBytes int64  `json:"memoryLimitsBytes,omitempty"`
	MemoryTotalBytes  int64  `json:"memoryTotalBytes,omitempty"`

	Since *time.Time `json:"since,omitempty"`

	// CanAct is whether the user may change what the item is about, beyond
	// looking at it: the screen it leads to lets them act.
	CanAct bool `json:"canAct"`
}

func TransformAttentionItem(item *attentionservice.Item, canAct bool) *AttentionItemResp {
	resp := &AttentionItemResp{
		Kind:              string(item.Kind),
		Severity:          string(item.Severity),
		Scope:             string(item.Scope.Type),
		Subject:           item.Subject,
		Running:           item.Running,
		Desired:           item.Desired,
		Restarts:          item.Restarts,
		LastError:         item.LastError,
		NodeState:         item.NodeState,
		MemoryLimitsBytes: item.MemoryLimits,
		MemoryTotalBytes:  item.MemoryTotal,
		CanAct:            canAct,
	}
	if item.Scope.Type == attentionservice.ScopeApp {
		resp.Project = &basedto.NamedObjectResp{ID: item.Scope.ProjectID, Name: item.ProjectName}
		resp.Env = item.Scope.ProjectEnv
		resp.App = &basedto.NamedObjectResp{ID: item.Scope.AppID, Name: item.Subject}
	}
	if !item.Since.IsZero() {
		since := item.Since
		resp.Since = &since
	}
	return resp
}
