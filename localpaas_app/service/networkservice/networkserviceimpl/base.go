package networkserviceimpl

import (
	"context"
	"time"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/infra/gocache"
	"github.com/localpaas/localpaas/services/docker"
)

const (
	cacheKeyGlobalRoutingNetworkID = "network:globalRoutingNetId"
	cacheExpGlobalRoutingNetworkID = 5 * time.Minute
)

func (s *service) FindGlobalRoutingNetworkID(ctx context.Context) (string, error) {
	if netID, _ := gocache.Global.GetStr(cacheKeyGlobalRoutingNetworkID); netID != "" {
		return netID, nil
	}

	net, err := s.dockerManager.NetworkList(ctx, func(options *network.ListOptions) {
		options.Filters = filters.NewArgs(filters.Arg("name", base.NetworkGlobalRouting))
	})
	if err != nil {
		return "", apperrors.Wrap(err)
	}

	var netID string
	if len(net) == 0 {
		netID, err = s.createGlobalRoutingNetwork(ctx)
		if err != nil {
			return "", apperrors.New(err).WithMsgLog("failed to create global routing network")
		}
	} else {
		netID = net[0].ID
	}

	// Cache the network ID
	_ = gocache.Global.Set(cacheKeyGlobalRoutingNetworkID, netID, cacheExpGlobalRoutingNetworkID)

	return netID, nil
}

func (s *service) createGlobalRoutingNetwork(ctx context.Context) (string, error) {
	resp, err := s.dockerManager.NetworkCreate(ctx, base.NetworkGlobalRouting, func(options *network.CreateOptions) {
		options.Driver = docker.NetworkDriverOverlay
		options.Scope = docker.NetworkScopeSwarm
		options.Attachable = true
		options.Labels = map[string]string{
			"localpaas.network.routing": "true",
		}
	})
	if err != nil {
		return "", apperrors.Wrap(err)
	}
	return resp.ID, nil
}
