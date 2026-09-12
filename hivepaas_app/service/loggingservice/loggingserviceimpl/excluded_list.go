package loggingserviceimpl

import (
	"context"

	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/loggingservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

// listExcludedApps loads every app and the app services, and names the ones
// logging will not show.
//
// One service listing, filtered to the stack namespace label every app service
// carries, rather than an inspect per app.
func (s *service) listExcludedApps(ctx context.Context, db database.IDB) ([]loggingservice.ExcludedApp, error) {
	apps, _, err := s.appRepo.List(ctx, db, "", nil)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	opts := client.ServiceListOptions{}
	docker.FilterAdd(&opts.Filters, "label", docker.StackLabelNamespace)
	resp, err := s.dockerManager.ServiceList(ctx, func(o *client.ServiceListOptions) { *o = opts })
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return excludedApps(apps, resp.Items), nil
}
