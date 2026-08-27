package agentserviceimpl

import (
	"context"
	"strconv"
	"strings"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/services/docker"
)

func (s *service) GetAgentAddrForNode(ctx context.Context, nodeID string) (string, error) {
	grpcPort := strconv.Itoa(config.Current.Agent.Port)
	if config.Current.DevMode.Enabled && config.Current.DevMode.ForceAgentLocal {
		return "localhost:" + grpcPort, nil
	}

	resp, err := s.dockerManager.TaskList(ctx, func(opts *client.TaskListOptions) {
		docker.FilterAdd(&opts.Filters, "service", base.HivepaasAgentServiceName)
		docker.FilterAdd(&opts.Filters, "node", nodeID)
		docker.FilterAdd(&opts.Filters, "desired-state", string(swarm.TaskStateRunning))
	})
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	if len(resp.Items) == 0 {
		return "", hperrors.Wrap(hperrors.ErrInfraNotFound).
			WithMsgLog("no running agent task found on node %s", nodeID)
	}

	var targetIP string
	for _, netAttachment := range resp.Items[0].NetworksAttachments {
		if netAttachment.Network.Spec.Name == base.NetworkHivepaasLocal {
			if len(netAttachment.Addresses) > 0 {
				addr := netAttachment.Addresses[0]
				if addr.IsValid() {
					targetIP = addr.Addr().String()
					break
				}
			}
		}
	}

	if targetIP == "" {
		return "", hperrors.Wrap(hperrors.ErrInfraNotFound).
			WithMsgLog("agent task on node %s is not connected to network %s", nodeID, base.NetworkHivepaasLocal)
	}

	return targetIP + ":" + grpcPort, nil
}

func (s *service) GetAgentAddrForNodeLabel(ctx context.Context, nodeLabel string) (string, error) {
	nodeLabel = strings.TrimSpace(nodeLabel)
	if nodeLabel == "" {
		return "", hperrors.Wrap(hperrors.ErrNodeWithLabelNotAvailable).WithParam("Label", nodeLabel)
	}

	parts := strings.SplitN(nodeLabel, "=", 2) //nolint:mnd
	key := strings.TrimSpace(parts[0])
	val := ""
	hasVal := len(parts) == 2 //nolint:mnd
	if hasVal {
		val = strings.TrimSpace(parts[1])
	}

	nodesResp, err := s.dockerManager.NodeList(ctx)
	if err != nil {
		return "", hperrors.Wrap(err)
	}

	var nodeID string
	for i := range nodesResp.Items {
		node := nodesResp.Items[i]
		if node.Status.State != swarm.NodeStateReady {
			continue
		}
		if actualVal, ok := node.Spec.Labels[key]; ok {
			if !hasVal || actualVal == val {
				nodeID = node.ID
				break
			}
		}
	}
	if nodeID == "" {
		return "", hperrors.Wrap(hperrors.ErrNodeWithLabelNotAvailable).WithParam("Label", nodeLabel)
	}
	return s.GetAgentAddrForNode(ctx, nodeID)
}
