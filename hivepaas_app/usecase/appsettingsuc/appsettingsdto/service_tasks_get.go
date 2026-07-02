package appsettingsdto

import (
	"sort"
	"time"

	"github.com/moby/moby/api/types/swarm"
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/copier"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/nodeuc/nodedto"
)

type GetAppServiceTasksReq struct {
	ProjectID string            `json:"-"`
	AppID     string            `json:"-"`
	States    []swarm.TaskState `json:"-" mapstructure:"state"`
}

func NewGetAppServiceTasksReq() *GetAppServiceTasksReq {
	return &GetAppServiceTasksReq{
		States: []swarm.TaskState{swarm.TaskStateRunning, swarm.TaskStateComplete, swarm.TaskStateShutdown,
			swarm.TaskStateFailed},
	}
}

func (req *GetAppServiceTasksReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetAppServiceTasksResp struct {
	Meta *basedto.Meta      `json:"meta"`
	Data []*ServiceTaskResp `json:"data"`
}

type ServiceTaskResp struct {
	ID           string                 `json:"id"`
	Slot         int                    `json:"slot"`
	Node         *nodedto.NodeBaseResp  `json:"node"`
	Status       *ServiceTaskStatusResp `json:"status"`
	DesiredState swarm.TaskState        `json:"desiredState"`
}

type ServiceTaskStatusResp struct {
	Timestamp       time.Time        `json:"timestamp"`
	State           swarm.TaskState  `json:"state"`
	Message         string           `json:"message"`
	Err             string           `json:"err"`
	ContainerStatus *ContainerStatus `json:"containerStatus"`
}

type ContainerStatus struct {
	ContainerID string `json:"containerId"`
}

func TransformServiceTask(task *swarm.Task, nodeMap map[string]*swarm.Node) (resp *ServiceTaskResp, err error) {
	if err = copier.Copy(&resp, task); err != nil {
		return nil, apperrors.New(err)
	}

	// Transform node details
	resp.Node = nodedto.TransformNodeBase(nodeMap[task.NodeID])

	return resp, nil
}

func TransformServiceTasks(tasks []swarm.Task, nodes []swarm.Node) (resp []*ServiceTaskResp, err error) {
	// Build node map
	nodeMap := make(map[string]*swarm.Node, len(nodes))
	for _, node := range nodes {
		nodeMap[node.ID] = &node
	}

	resp = make([]*ServiceTaskResp, 0, len(tasks))
	for i := range tasks {
		taskResp, err := TransformServiceTask(&tasks[i], nodeMap)
		if err != nil {
			return nil, apperrors.New(err)
		}
		resp = append(resp, taskResp)
	}

	//nolint:mnd,exhaustive
	stateRank := func(state swarm.TaskState) int {
		switch state {
		case swarm.TaskStateRunning:
			return 0
		case swarm.TaskStateComplete:
			return 1
		case swarm.TaskStateShutdown:
			return 2
		case swarm.TaskStateFailed:
			return 3
		default:
			return 4
		}
	}

	sort.Slice(resp, func(i, j int) bool {
		if resp[i].DesiredState != resp[j].DesiredState {
			return stateRank(resp[i].DesiredState) < stateRank(resp[j].DesiredState)
		}

		var stateI, stateJ swarm.TaskState
		var timeI, timeJ time.Time
		if resp[i].Status != nil {
			stateI = resp[i].Status.State
			timeI = resp[i].Status.Timestamp
		}
		if resp[j].Status != nil {
			stateJ = resp[j].Status.State
			timeJ = resp[j].Status.Timestamp
		}

		if stateI != stateJ {
			return stateRank(stateI) < stateRank(stateJ)
		}
		if resp[i].Slot != resp[j].Slot {
			return resp[i].Slot < resp[j].Slot
		}
		return timeI.After(timeJ)
	})

	return resp, nil
}
