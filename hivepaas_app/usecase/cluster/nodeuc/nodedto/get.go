package nodedto

import (
	"github.com/moby/moby/api/types/swarm"
	vld "github.com/tiendc/go-validator"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/copier"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/dockerhelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/services/docker"
)

const (
	nodeNameMaxLen = 100
)

type GetNodeReq struct {
	settings.GetSettingReq
}

func NewGetNodeReq() *GetNodeReq {
	return &GetNodeReq{}
}

func (req *GetNodeReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.GetSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetNodeResp struct {
	Meta *basedto.Meta `json:"meta"`
	Data *NodeResp     `json:"data"`
}

type NodeResp struct {
	*settings.BaseSettingResp

	Labels       map[string]string       `json:"labels"`
	Hostname     string                  `json:"hostname"`
	Addr         string                  `json:"addr"`
	State        docker.NodeState        `json:"state"`
	Availability docker.NodeAvailability `json:"availability"`
	Role         docker.NodeRole         `json:"role"`
	IsLeader     bool                    `json:"isLeader"`
	Platform     *NodePlatformResp       `json:"platform"`
	Resources    *NodeResources          `json:"resources"`
	EngineDesc   *NodeEngineDescResp     `json:"engineDesc"`
}

type NodeBaseResp struct {
	ID           string                  `json:"id"`
	Name         string                  `json:"name"`
	Hostname     string                  `json:"hostname"`
	Addr         string                  `json:"addr"`
	State        docker.NodeState        `json:"state"`
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

func TransformNode(
	setting *entity.Setting,
	_ *entity.RefObjects,
	refClusterObjects *entity.RefClusterObjects,
	detailed bool,
) (resp *NodeResp, err error) {
	nodeEnt := setting.MustAsClusterNode()
	if err = copier.Copy(&resp, nodeEnt); err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp.BaseSettingResp, err = settings.TransformSettingBase(setting)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	node := refClusterObjects.RefNodes[dockerhelper.ParseID(setting.ID)]

	resp.Name = gofn.Coalesce(node.Spec.Name, "<unset>")
	resp.State = docker.NodeState(node.Status.State)
	resp.Availability = docker.NodeAvailability(node.Spec.Availability)
	resp.Role = docker.NodeRole(node.Spec.Role)
	isManager := node.Spec.Role == swarm.NodeRoleManager
	resp.IsLeader = isManager && node.ManagerStatus != nil && node.ManagerStatus.Leader
	resp.Hostname = node.Description.Hostname
	resp.Addr = node.Status.Addr
	resp.Platform = &NodePlatformResp{
		Architecture: node.Description.Platform.Architecture,
		OS:           node.Description.Platform.OS,
	}
	resp.Resources = &NodeResources{
		CPUs:        node.Description.Resources.NanoCPUs / docker.UnitCPUNano,
		Memory:      unit.DataSize(node.Description.Resources.MemoryBytes),
		MemoryBytes: node.Description.Resources.MemoryBytes,
	}
	resp.CreatedAt = node.CreatedAt
	resp.UpdatedAt = node.UpdatedAt

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
	return resp, nil
}

func TransformNodeBase(node *swarm.Node) *NodeBaseResp {
	if node == nil {
		return nil
	}
	isManager := node.Spec.Role == swarm.NodeRoleManager
	return &NodeBaseResp{
		ID:           dockerhelper.WrapNodeID(node.ID),
		Name:         node.Spec.Name,
		State:        docker.NodeState(node.Status.State),
		Availability: docker.NodeAvailability(node.Spec.Availability),
		Role:         docker.NodeRole(node.Spec.Role),
		IsLeader:     isManager && node.ManagerStatus != nil && node.ManagerStatus.Leader,
		Hostname:     node.Description.Hostname,
		Addr:         node.Status.Addr,
	}
}
