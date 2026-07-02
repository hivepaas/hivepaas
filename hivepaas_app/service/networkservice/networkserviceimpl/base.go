package networkserviceimpl

import (
	"context"
	"time"

	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/gocache"
	"github.com/hivepaas/hivepaas/services/docker"
)

const (
	cacheKeyGlobalRoutingNetworkID = "network:global-net-id"
	cacheExpGlobalRoutingNetworkID = 5 * time.Minute
)

func (s *service) GetGlobalRoutingNetworkID(ctx context.Context) (string, error) {
	if netID, _ := gocache.Global.GetStr(cacheKeyGlobalRoutingNetworkID); netID != "" {
		return netID, nil
	}

	listResp, err := s.dockerManager.NetworkList(ctx, func(opts *client.NetworkListOptions) {
		docker.FilterAdd(&opts.Filters, "name", base.NetworkGlobalRouting)
	})
	if err != nil {
		return "", apperrors.New(err)
	}

	var netID string
	if len(listResp.Items) == 0 {
		netID, err = s.createGlobalRoutingNetwork(ctx)
		if err != nil {
			return "", apperrors.New(err).WithMsgLog("failed to create global routing network")
		}
	} else {
		netID = listResp.Items[0].ID
	}

	// Cache the network ID
	_ = gocache.Global.Set(cacheKeyGlobalRoutingNetworkID, netID, cacheExpGlobalRoutingNetworkID)

	return netID, nil
}

func (s *service) createGlobalRoutingNetwork(ctx context.Context) (string, error) {
	resp, err := s.dockerManager.NetworkCreate(ctx, base.NetworkGlobalRouting,
		func(options *client.NetworkCreateOptions) {
			options.Driver = docker.NetworkDriverOverlay
			options.Scope = docker.NetworkScopeSwarm
			options.Attachable = true
			options.Labels = map[string]string{
				"hivepaas.network.routing": "true",
			}
		})
	if err != nil {
		return "", apperrors.New(err)
	}
	return resp.ID, nil
}
