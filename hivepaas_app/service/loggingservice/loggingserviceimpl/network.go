package loggingserviceimpl

import (
	"context"

	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/services/docker"
)

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
	return hperrors.Wrap(hperrors.ErrLoggingAPINetworkMissing).WithParam("Name", base.NetworkHivepaasLocal)
}
