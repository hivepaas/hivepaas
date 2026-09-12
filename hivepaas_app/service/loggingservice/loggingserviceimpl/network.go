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

// checkInternalNetwork makes sure the network the API is on exists, so that a
// backend is not deployed where nothing can query it.
//
// The name is a constant rather than something read off the API's own service:
// every other name in this codebase - hivepaas_app, hivepaas_db, hivepaas_net -
// already assumes the stack is deployed under that one name, so deriving this
// one would be a mechanism nothing else needs.
func (s *service) checkInternalNetwork(ctx context.Context) error {
	list, err := s.dockerManager.NetworkList(ctx, func(opts *client.NetworkListOptions) {
		docker.FilterAdd(&opts.Filters, "name", base.NetworkHivepaasLocal)
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	// The name filter matches substrings.
	for i := range list.Items {
		if list.Items[i].Name == base.NetworkHivepaasLocal {
			return nil
		}
	}
	return hperrors.Wrap(loggingservice.ErrAPINetworkMissing).WithParam("Name", base.NetworkHivepaasLocal)
}
