package loggingserviceimpl

import (
	"context"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/loggingservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

// Status reports the apps HivePaaS runs for logging and what they are doing.
//
// It never fails the settings screen over a service it cannot read: an app whose
// service is missing is reported with no tasks, which is what it has.
func (s *service) Status(
	ctx context.Context,
	db database.IDB,
	setting *entity.Setting,
) (status *loggingservice.Status, err error) {
	status = &loggingservice.Status{BackendResources: defaultBackendResources()}
	if setting == nil {
		if setting, err = s.loadSettings(ctx, db, false); err != nil {
			return nil, hperrors.Wrap(err)
		}
	}
	if setting != nil {
		cfg, err := setting.AsLoggingSettings()
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		status.Enabled = cfg.Enabled
	}

	backend, err := s.systemAppService.LoadApp(ctx, db, backendAppKey)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if backend != nil {
		var svc *swarm.Service
		status.Backend, svc = s.appStatus(ctx, backend)
		// The limits are read off the service, where they live: see
		// applyBackendResources.
		if svc != nil {
			status.BackendResources = resourcesOf(&svc.Spec)
		}
	}

	collector, err := s.systemAppService.LoadApp(ctx, db, collectorAppKey)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if collector != nil {
		status.Collector, _ = s.appStatus(ctx, collector)
	}
	return status, nil
}

// appStatus reads the app's service with its task counts, which only a listing
// carries: an inspect of the same service has none.
func (s *service) appStatus(ctx context.Context, app *entity.App) (*loggingservice.AppStatus, *swarm.Service) {
	status := &loggingservice.AppStatus{AppID: app.ID, ProjectID: app.ProjectID}
	if app.ProjectEnv != nil {
		status.ProjectEnv = app.ProjectEnv.Key
	}
	if app.ServiceID == "" {
		return status, nil
	}
	listed, err := s.dockerManager.ServiceList(ctx, func(opts *client.ServiceListOptions) {
		docker.FilterAdd(&opts.Filters, "id", app.ServiceID)
		opts.Status = true
	})
	if err != nil {
		s.logger.Warnf("cannot read the service of logging app %s: %v", app.Key, err)
		return status, nil
	}
	// The id filter matches prefixes.
	for i := range listed.Items {
		svc := &listed.Items[i]
		if svc.ID != app.ServiceID {
			continue
		}
		if svc.ServiceStatus != nil {
			status.RunningTasks = svc.ServiceStatus.RunningTasks
			status.DesiredTasks = svc.ServiceStatus.DesiredTasks
		}
		return status, svc
	}
	return status, nil
}
