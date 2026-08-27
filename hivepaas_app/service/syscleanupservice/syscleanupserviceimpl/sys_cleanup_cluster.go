package syscleanupserviceimpl

import (
	"context"
	"errors"
	"fmt"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/client/nodecleanupservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clustercleanupservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/nodecleanupagentuc/nodecleanupagentdto"
)

//nolint:gocognit
func (s *service) sysCleanupCluster(
	ctx context.Context,
	data *sysCleanupData,
) (err error) {
	clusterCleanup := &data.SysCleanupSettings.ClusterCleanup
	if !clusterCleanup.Enabled {
		return nil
	}

	if data.TaskOutput.ClusterCleanup == nil {
		data.TaskOutput.ClusterCleanup = &entity.ClusterCleanupOutput{}
	}

	// 1. Cleanup cluster on current node (manager node)
	managerNodeID, _ := s.dockerManager.NodeCurrentID(ctx)

	resp, e := s.clusterCleanupService.Cleanup(ctx, &clustercleanupservice.ClusterCleanupReq{
		TaskExecData:      data.TaskExecData,
		CleanupSettings:   clusterCleanup,
		CleanupContainers: data.CleanupClusterContainers,
		CleanupImages:     data.CleanupClusterImages,
		CleanupVolumes:    data.CleanupClusterVolumes,
		CleanupNetworks:   data.CleanupClusterNetworks,
		CleanupBuildCache: data.CleanupClusterBuildCache,
	})
	if e != nil {
		err = errors.Join(err, hperrors.Wrap(e))
	}
	if resp != nil && resp.Output != nil {
		if resp.Output.NodeID == "" {
			resp.Output.NodeID = managerNodeID
		}
		if resp.Output.NodeName == "" {
			resp.Output.NodeName = "<manager>"
		}
		data.TaskOutput.ClusterCleanup.Nodes = append(data.TaskOutput.ClusterCleanup.Nodes, resp.Output)
	}

	// 2. Scan all Swarm nodes and run cleanup via gRPC for other nodes
	nodesResp, listErr := s.dockerManager.NodeList(ctx)
	if listErr != nil {
		return errors.Join(err, hperrors.Wrap(listErr))
	}

	for i := range nodesResp.Items {
		node := &nodesResp.Items[i]
		if node.ID == managerNodeID {
			// Already cleaned up above
			continue
		}

		nodeName := node.Spec.Name
		if nodeName == "" {
			nodeName = node.Description.Hostname
		}

		// Skip if node is down or disconnected
		if node.Status.State == swarm.NodeStateDown || node.Status.State == swarm.NodeStateDisconnected {
			continue
		}

		agentAddr, addrErr := s.agentService.GetAgentAddrForNode(ctx, node.ID)
		if addrErr != nil {
			nodeOutput := &entity.ClusterNodeCleanupOutput{
				NodeID:               node.ID,
				NodeName:             nodeName,
				ContainersPruneError: fmt.Sprintf("Failed to get agent address: %v", addrErr),
			}
			data.TaskOutput.ClusterCleanup.Nodes = append(data.TaskOutput.ClusterCleanup.Nodes, nodeOutput)
			continue
		}

		agentClient, clientErr := nodecleanupservice.NewNodeCleanupServiceClient(agentAddr)
		if clientErr != nil {
			nodeOutput := &entity.ClusterNodeCleanupOutput{
				NodeID:               node.ID,
				NodeName:             nodeName,
				ContainersPruneError: fmt.Sprintf("Failed to connect to agent at %s: %v", agentAddr, clientErr),
			}
			data.TaskOutput.ClusterCleanup.Nodes = append(data.TaskOutput.ClusterCleanup.Nodes, nodeOutput)
			continue
		}

		taskID := ""
		if data.Task != nil {
			taskID = data.Task.ID
		}

		cleanupResp, cleanupErr := agentClient.NodeCleanup(ctx, &nodecleanupagentdto.NodeCleanupReq{
			TaskID: taskID,
			ClusterCleanupReq: clustercleanupservice.ClusterCleanupReq{
				CleanupSettings:   clusterCleanup,
				CleanupContainers: data.CleanupClusterContainers,
				CleanupImages:     data.CleanupClusterImages,
				CleanupVolumes:    data.CleanupClusterVolumes,
				CleanupNetworks:   data.CleanupClusterNetworks,
				CleanupBuildCache: data.CleanupClusterBuildCache,
			},
		})
		_ = agentClient.Close()

		if cleanupErr != nil {
			nodeOutput := &entity.ClusterNodeCleanupOutput{
				NodeID:               node.ID,
				NodeName:             nodeName,
				ContainersPruneError: fmt.Sprintf("Remote node cleanup error: %v", cleanupErr),
			}
			data.TaskOutput.ClusterCleanup.Nodes = append(data.TaskOutput.ClusterCleanup.Nodes, nodeOutput)
			err = errors.Join(err, hperrors.Wrap(cleanupErr))
			continue
		}

		if cleanupResp != nil {
			nodeOutput := &cleanupResp.ClusterNodeCleanupOutput
			if nodeOutput.NodeID == "" {
				nodeOutput.NodeID = node.ID
			}
			if nodeOutput.NodeName == "" {
				nodeOutput.NodeName = nodeName
			}
			data.TaskOutput.ClusterCleanup.Nodes = append(data.TaskOutput.ClusterCleanup.Nodes, nodeOutput)
		}
	}

	return hperrors.Wrap(err)
}
