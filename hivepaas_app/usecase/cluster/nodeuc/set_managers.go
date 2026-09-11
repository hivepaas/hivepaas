package nodeuc

import (
	"context"
	"fmt"
	"sort"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/auditdetail"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/entityutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/clusteraudit"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/nodeuc/nodedto"
)

func (uc *UC) SetManagerNodes(
	ctx context.Context,
	auth *basedto.Auth,
	req *nodedto.SetManagerNodesReq,
) (*nodedto.SetManagerNodesResp, error) {
	dbNodeSettings, _, err := uc.SettingRepo.List(ctx, uc.DB, nil, nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeClusterNode),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	dbNodeMap := entityutil.SliceToIDMap(dbNodeSettings)

	listResp, err := uc.dockerManager.NodeList(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	nodes := listResp.Items

	existingNodes := make(map[string]*swarm.Node)
	for i := range nodes {
		existingNodes[nodes[i].ID] = &nodes[i]
	}

	targetManagerIDs := make(map[string]bool)
	for _, nodeReq := range req.Nodes {
		dbNode := dbNodeMap[nodeReq.ID]
		var dockerNode *swarm.Node
		if dbNode != nil {
			dockerNode = existingNodes[dbNode.RefID]
		}
		if dockerNode == nil {
			return nil, hperrors.NewNotFound(fmt.Sprintf("Node %v", nodeReq.ID))
		}
		targetManagerIDs[dockerNode.ID] = true
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

	// Recorded before anything moves. Promotion and demotion go straight to
	// Docker, one node at a time, and a failure halfway leaves the cluster in a
	// state no rollback here can undo - so the only order that cannot lose the
	// record is to write it first. Who holds manager role is who controls the
	// cluster, which is why this is worth a record at all.
	err = clusteraudit.Record(ctx, uc.AuditService, uc.DB, auth,
		clusteraudit.Target{ResType: base.ResourceTypeClusterNode}, "node-set-managers",
		auditdetail.New().
			Set("promoting", nodeIDs(promoteNodes)).
			Set("demoting", nodeIDs(demoteNodes)))
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	// 1. Promote worker nodes to managers first to preserve quorum
	for _, node := range promoteNodes {
		inspect, err := uc.dockerManager.NodeInspect(ctx, node.ID)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		latestNode := &inspect.Node
		spec := latestNode.Spec
		spec.Role = swarm.NodeRoleManager
		_, err = uc.dockerManager.NodeUpdate(ctx, latestNode.ID, &latestNode.Version, &spec)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
	}

	// 2. Demote manager nodes to workers
	for _, node := range uc.sortNodesToDemote(ctx, demoteNodes) {
		inspect, err := uc.dockerManager.NodeInspect(ctx, node.ID)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		latestNode := &inspect.Node
		spec := latestNode.Spec
		spec.Role = swarm.NodeRoleWorker
		_, err = uc.dockerManager.NodeUpdate(ctx, latestNode.ID, &latestNode.Version, &spec)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
	}

	// 3. Sync nodes back to the database
	_, err = uc.clusterService.SyncNodes(ctx, uc.db)
	if err != nil {
		return nil, hperrors.Wrap(err)
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

// nodeIDs names the nodes an operation is about to move, so the entry survives
// the operation failing partway through.
func nodeIDs(nodes []*swarm.Node) []string {
	ids := make([]string, 0, len(nodes))
	for _, node := range nodes {
		ids = append(ids, node.ID)
	}
	return ids
}
