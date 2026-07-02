package agentserviceimpl

import (
	"context"
	"strconv"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
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
		return "", apperrors.New(err)
	}
	if len(resp.Items) == 0 {
		return "", apperrors.New(apperrors.ErrInfraNotFound).
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
		return "", apperrors.New(apperrors.ErrInfraNotFound).
			WithMsgLog("agent task on node %s is not connected to network %s", nodeID, base.NetworkHivepaasLocal)
	}

	return targetIP + ":" + grpcPort, nil
}
