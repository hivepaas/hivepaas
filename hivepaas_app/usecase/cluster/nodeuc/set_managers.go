package nodeuc

import (
	"context"
	"fmt"
	"sort"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/dockerhelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/nodeuc/nodedto"
)

func (uc *UC) SetManagerNodes(
	ctx context.Context,
	auth *basedto.Auth,
	req *nodedto.SetManagerNodesReq,
) (*nodedto.SetManagerNodesResp, error) {
	listResp, err := uc.dockerManager.NodeList(ctx)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	nodes := listResp.Items

	existingNodes := make(map[string]*swarm.Node)
	for i := range nodes {
		existingNodes[nodes[i].ID] = &nodes[i]
	}

	targetManagerIDs := make(map[string]bool)
	for _, nodeReq := range req.Nodes {
		nodeID := dockerhelper.ParseID(nodeReq.ID)
		if _, ok := existingNodes[nodeID]; !ok {
			return nil, apperrors.NewNotFound(fmt.Sprintf("Node %v", nodeReq.ID))
		}
		targetManagerIDs[nodeID] = true
	}

	var promoteNodes []*swarm.Node
	var demoteNodes []*swarm.Node

	for i := range nodes {
		node := &nodes[i]
		inTarget := targetManagerIDs[node.ID]
		isManager := node.Spec.Role == swarm.NodeRoleManager

		if inTarget && !isManager {
			promoteNodes = append(promoteNodes, node)
		} else if !inTarget && isManager {
			demoteNodes = append(demoteNodes, node)
		}
	}

	// 1. Promote worker nodes to managers first to preserve quorum
	for _, node := range promoteNodes {
		inspect, err := uc.dockerManager.NodeInspect(ctx, node.ID)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		latestNode := &inspect.Node
		spec := latestNode.Spec
		spec.Role = swarm.NodeRoleManager
		_, err = uc.dockerManager.NodeUpdate(ctx, latestNode.ID, &latestNode.Version, &spec)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
	}

	// 2. Demote manager nodes to workers
	for _, node := range uc.sortNodesToDemote(ctx, demoteNodes) {
		inspect, err := uc.dockerManager.NodeInspect(ctx, node.ID)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		latestNode := &inspect.Node
		spec := latestNode.Spec
		spec.Role = swarm.NodeRoleWorker
		_, err = uc.dockerManager.NodeUpdate(ctx, latestNode.ID, &latestNode.Version, &spec)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
	}

	// 3. Sync nodes back to the database
	_, err = uc.clusterService.SyncNodes(ctx, uc.db)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &nodedto.SetManagerNodesResp{}, nil
}

func (uc *UC) sortNodesToDemote(
	ctx context.Context,
	demoteNodes []*swarm.Node,
) []*swarm.Node {
	// Sort demoteNodes to ensure the current node and the leader manager node are demoted last.
	// This maintains Docker Swarm quorum stability and avoids interrupting the API client connection.
	currNodeID, _ := uc.dockerManager.NodeCurrentID(ctx)
	sort.SliceStable(demoteNodes, func(i, j int) bool {
		nodeI := demoteNodes[i]
		nodeJ := demoteNodes[j]

		isSpecialI := nodeI.ID == currNodeID || (nodeI.ManagerStatus != nil && nodeI.ManagerStatus.Leader)
		isSpecialJ := nodeJ.ID == currNodeID || (nodeJ.ManagerStatus != nil && nodeJ.ManagerStatus.Leader)

		if isSpecialI && !isSpecialJ {
			return false
		}
		if !isSpecialI && isSpecialJ {
			return true
		}

		if isSpecialI && isSpecialJ {
			isLeaderI := nodeI.ManagerStatus != nil && nodeI.ManagerStatus.Leader
			isLeaderJ := nodeJ.ManagerStatus != nil && nodeJ.ManagerStatus.Leader
			if isLeaderI && !isLeaderJ {
				return false
			}
			if !isLeaderI && isLeaderJ {
				return true
			}
		}
		return false
	})
	return demoteNodes
}
