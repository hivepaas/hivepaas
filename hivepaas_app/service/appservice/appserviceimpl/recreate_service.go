package appserviceimpl

import (
	"context"
	"time"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clusterservice"
)

const (
	// How long to wait for the old service tasks to terminate before recreating the service under
	// the same name. Swarm frees the service name only once the service object is gone.
	recreateServiceStopCheckInterval = time.Second
)

// RecreateServiceWithSpec deletes the app swarm service and creates it again from newSpec.
//
// Swarm refuses to change the service mode variant (replicated <-> global <-> jobs) on an existing
// service, so switching mode is only possible by recreating it. The service name is reused, which
// forces the deletion to complete before the creation starts.
//
// This causes downtime: every task of the app is stopped before the new service is created. Callers
// must have confirmed that with the user beforehand.
//
// On failure to create the new service, the original spec is recreated so the app keeps running.
// If that recovery also fails the app is left without a service; the error says so explicitly and
// the app can be restored by redeploying it.
//
// The caller is responsible for persisting the returned service ID onto the app.
func (s *service) RecreateServiceWithSpec(
	ctx context.Context,
	app *entity.App,
	oldSpec *swarm.ServiceSpec,
	newSpec *swarm.ServiceSpec,
) (string, error) {
	if app == nil || app.ServiceID == "" {
		return "", hperrors.Wrap(hperrors.ErrAppNotFound)
	}
	if oldSpec == nil || newSpec == nil {
		return "", hperrors.Wrap(hperrors.ErrInfraInternal).
			WithNTParam("Error", "missing service spec to recreate")
	}

	oldServiceID := app.ServiceID

	if err := s.clusterService.ServiceRemove(ctx, oldServiceID,
		clusterservice.ItemRemovalRetryMax, 0); err != nil {
		return "", hperrors.Wrap(err)
	}

	// The name stays taken until the tasks are gone, so creating too early fails on a name conflict.
	if _, err := s.dockerManager.ServiceWaitUntilStopped(ctx, oldServiceID,
		recreateServiceStopCheckInterval); err != nil {
		return "", hperrors.Wrap(err)
	}

	res, err := s.dockerManager.ServiceCreate(ctx, newSpec)
	if err == nil && res.ID == "" { // should never happen
		err = hperrors.Wrap(hperrors.ErrInfraInternal).
			WithNTParam("Error", "empty service ID returned")
	}
	if err != nil {
		return "", s.rollbackRecreatedService(ctx, app, oldSpec, err)
	}

	return res.ID, nil
}

// rollbackRecreatedService puts the original service back after a failed recreation, and returns
// the error to report to the caller.
func (s *service) rollbackRecreatedService(
	ctx context.Context,
	app *entity.App,
	oldSpec *swarm.ServiceSpec,
	createErr error,
) error {
	restored, restoreErr := s.dockerManager.ServiceCreate(ctx, oldSpec)
	if restoreErr != nil || restored == nil || restored.ID == "" {
		// The app now has no service at all. Surface it loudly: no automatic recovery exists.
		return hperrors.Wrap(createErr).
			WithNTParam("Error", "failed to change the service mode and to restore the previous "+
				"service; the app currently has no running service and must be redeployed")
	}

	// Keep the app pointing at the service that actually exists now.
	app.ServiceID = restored.ID

	return hperrors.Wrap(createErr).
		WithNTParam("Error", "failed to change the service mode; the previous service was restored")
}
