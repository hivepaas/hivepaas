package nodedto

import (
	"time"

	"github.com/moby/moby/api/types/swarm"
	vld "github.com/tiendc/go-validator"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/services/docker"
)

const (
	nodeIDMaxLen   = 100
	nodeNameMaxLen = 100
)

type GetNodeReq struct {
	NodeID string `json:"-"`
}

func NewGetNodeReq() *GetNodeReq {
	return &GetNodeReq{}
}

func (req *GetNodeReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	// NOTE: node id is docker id, it's not ULID
	validators = append(validators, basedto.ValidateStr(&req.NodeID, true, 1, nodeIDMaxLen, "nodeId")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetNodeResp struct {
	Meta *basedto.Meta `json:"meta"`
	Data *NodeResp     `json:"data"`
}

type NodeResp struct {
	ID           string                  `json:"id"`
	Name         string                  `json:"name"`
	Labels       map[string]string       `json:"labels"`
	Hostname     string                  `json:"hostname"`
	Addr         string                  `json:"addr"`
	Status       docker.NodeStatus       `json:"status"`
	Availability docker.NodeAvailability `json:"availability"`
	Role         docker.NodeRole         `json:"role"`
	IsLeader     bool                    `json:"isLeader"`
	Platform     *NodePlatformResp       `json:"platform"`
	Resources    *NodeResources          `json:"resources"`
	EngineDesc   *NodeEngineDescResp     `json:"engineDesc"`
	UpdateVer    int                     `json:"updateVer"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type NodeBaseResp struct {
	ID           string                  `json:"id"`
	Name         string                  `json:"name"`
	Hostname     string                  `json:"hostname"`
	Addr         string                  `json:"addr"`
	Status       docker.NodeStatus       `json:"status"`
	Availability docker.NodeAvailability `json:"availability"`
	Role         docker.NodeRole         `json:"role"`
	IsLeader     bool                    `json:"isLeader"`
}

type NodePlatformResp struct {
	Architecture string `json:"architecture"`
	OS           string `json:"os"`
}

type NodeResources struct {
	CPUs        int64         `json:"cpus"`
	Memory      unit.DataSize `json:"memory"`
	MemoryBytes int64         `json:"memoryBytes"`
}

type NodeEngineDescResp struct {
	EngineVersion string                `json:"engineVersion"`
	Labels        map[string]string     `json:"labels"`
	Plugins       []*NodePluginDescResp `json:"plugins,omitempty"`
}

type NodePluginDescResp struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

func TransformNode(node *swarm.Node, detailed bool) *NodeResp {
	isManager := node.Spec.Role == swarm.NodeRoleManager
	resp := &NodeResp{
		ID:           node.ID,
		Name:         gofn.Coalesce(node.Spec.Name, "<unset>"),
		Status:       docker.NodeStatus(node.Status.State),
		Availability: docker.NodeAvailability(node.Spec.Availability),
		Role:         docker.NodeRole(node.Spec.Role),
		IsLeader:     isManager && node.ManagerStatus != nil && node.ManagerStatus.Leader,
		Hostname:     node.Description.Hostname,
		Addr:         node.Status.Addr,
		Platform: &NodePlatformResp{
			Architecture: node.Description.Platform.Architecture,
			OS:           node.Description.Platform.OS,
		},
		Resources: &NodeResources{
			CPUs:        node.Description.Resources.NanoCPUs / docker.UnitCPUNano,
			Memory:      unit.DataSize(node.Description.Resources.MemoryBytes),
			MemoryBytes: node.Description.Resources.MemoryBytes,
		},
		UpdateVer: int(node.Version.Index), //nolint:gosec
		CreatedAt: node.CreatedAt,
		UpdatedAt: node.UpdatedAt,
	}
	if detailed {
		resp.Labels = node.Spec.Labels
		resp.EngineDesc = &NodeEngineDescResp{
			EngineVersion: node.Description.Engine.EngineVersion,
			Labels:        node.Description.Engine.Labels,
			Plugins: gofn.MapSlice(node.Description.Engine.Plugins, func(p swarm.PluginDescription) *NodePluginDescResp {
				return &NodePluginDescResp{
					Type: p.Type,
					Name: p.Name,
				}
			}),
		}
	}
	return resp
}

func TransformNodeBase(node *swarm.Node) *NodeBaseResp {
	if node == nil {
		return nil
	}
	isManager := node.Spec.Role == swarm.NodeRoleManager
	return &NodeBaseResp{
		ID:           node.ID,
		Name:         node.Spec.Name,
		Status:       docker.NodeStatus(node.Status.State),
		Availability: docker.NodeAvailability(node.Spec.Availability),
		Role:         docker.NodeRole(node.Spec.Role),
		IsLeader:     isManager && node.ManagerStatus != nil && node.ManagerStatus.Leader,
		Hostname:     node.Description.Hostname,
		Addr:         node.Status.Addr,
	}
}
