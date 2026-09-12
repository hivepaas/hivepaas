package loggingserviceimpl

import (
	"context"

	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/loggingservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

// ensureLoggingNetwork returns the overlay the collector and the backend share,
// creating it the first time.
//
// Swarm services resolve each other by name only across a network they both
// join. Without one, vlagent's writes to hivepaas-victoria-logs fail to
// resolve and nothing is ever stored.
//
// Not attachable: a container started by hand has no business on it.
func (s *service) ensureLoggingNetwork(ctx context.Context) (string, error) {
	list, err := s.dockerManager.NetworkList(ctx, func(opts *client.NetworkListOptions) {
		docker.FilterAdd(&opts.Filters, "name", NetworkLogging)
	})
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	// The name filter matches substrings.
	for i := range list.Items {
		if list.Items[i].Name == NetworkLogging {
			return list.Items[i].ID, nil
		}
	}

	resp, err := s.dockerManager.NetworkCreate(ctx, NetworkLogging, func(opts *client.NetworkCreateOptions) {
		opts.Driver = docker.NetworkDriverOverlay
		opts.Scope = docker.NetworkScopeSwarm
		opts.Attachable = false
		opts.Options = map[string]string{docker.NetworkOptionDriverMTU: docker.DefaultOverlayNetworkMTU}
		opts.Labels = map[string]string{LabelManagedBy: "true"}
	})
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	return resp.ID, nil
}

// apiPrivateNetworks is every network HivePaaS's own API is on except the
// routing network.
//
// The backend joins these so that the API can query it. The routing network is
// left out because every publicly exposed app is on it, and the backend has no
// authentication.
func (s *service) apiPrivateNetworks(ctx context.Context) ([]string, error) {
	svc, err := s.hpAppService.GetHpAppSwarmService(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	routingID, err := s.networkService.GetGlobalRoutingNetworkID(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	out := []string{}
	for _, n := range svc.Spec.TaskTemplate.Networks {
		if n.Target == routingID || n.Target == base.NetworkGlobalRouting {
			continue
		}
		out = append(out, n.Target)
	}
	if len(out) == 0 {
		return nil, hperrors.Wrap(loggingservice.ErrAPINetworkMissing)
	}
	return out, nil
}
