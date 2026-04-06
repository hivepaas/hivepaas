package networkservice

import (
	"context"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/services/docker"
)

const (
	GlobalRoutingNetwork = "localpaas_net"
)

var (
	GlobalRoutingNetworkID = "" // cache value
)

func (s *networkService) FindGlobalRoutingNetworkID(ctx context.Context) (string, error) {
	// TODO: do we need a lock?
	if GlobalRoutingNetworkID != "" {
		return GlobalRoutingNetworkID, nil
	}

	net, err := s.dockerManager.NetworkList(ctx, func(options *network.ListOptions) {
		options.Filters = filters.NewArgs(filters.Arg("name", GlobalRoutingNetwork))
	})
	if err != nil {
		return "", apperrors.Wrap(err)
	}

	if len(net) == 0 {
		err = s.createGlobalRoutingNetwork(ctx)
		if err != nil {
			return "", apperrors.New(err).WithMsgLog("failed to create global routing network")
		}
	} else {
		GlobalRoutingNetworkID = net[0].ID
	}

	return GlobalRoutingNetworkID, nil
}

func (s *networkService) createGlobalRoutingNetwork(ctx context.Context) error {
	resp, err := s.dockerManager.NetworkCreate(ctx, GlobalRoutingNetwork, func(options *network.CreateOptions) {
		options.Driver = docker.NetworkDriverOverlay
		options.Scope = docker.NetworkScopeSwarm
		options.Attachable = true
		options.Labels = map[string]string{
			"localpaas.network.routing": "true",
		}
	})
	if err != nil {
		return apperrors.Wrap(err)
	}

	GlobalRoutingNetworkID = resp.ID
	return nil
}
