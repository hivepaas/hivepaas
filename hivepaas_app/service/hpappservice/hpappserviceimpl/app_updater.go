package hpappserviceimpl

import (
	"context"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// serviceUpdateRetryMax is how many times a rejected service update is retried.
//
// ServiceUpdateFunc re-inspects the service before each attempt, so the change
// is re-applied to a fresh Version - and a stale Version is what swarm rejects
// when something else has written to the service in between.
const serviceUpdateRetryMax = 2

func (s *service) GetHpUpdaterSwarmService(ctx context.Context) (*swarm.Service, error) {
	service, err := s.dockerManager.ServiceGetByName(ctx, base.HivepaasUpdaterServiceName, false)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return service, nil
}

func (s *service) RestartHpUpdaterSwarmService(ctx context.Context) error {
	service, err := s.GetHpUpdaterSwarmService(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}

	service.Spec.TaskTemplate.ForceUpdate++
	_, err = s.dockerManager.ServiceUpdate(ctx, service.ID, &service.Version, &service.Spec)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (s *service) ShutdownHpUpdaterSwarmService(ctx context.Context) error {
	service, err := s.GetHpUpdaterSwarmService(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}

	if service.Spec.Mode.Replicated == nil || *service.Spec.Mode.Replicated.Replicas == 0 {
		return nil
	}
	service.Spec.Mode.Replicated.Replicas = new(uint64(0))

	_, err = s.dockerManager.ServiceUpdate(ctx, service.ID, &service.Version, &service.Spec)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

// EnsureHpAppRunning brings the main app back if it is scaled to zero.
//
// It exists for one situation. A system update stops the app before it touches
// anything, and puts the replicas back when it finishes - successfully or not.
// If the updater itself dies in between, killed or timed out, nothing puts them
// back: the dashboard, the API and the task queue all need the app to be running,
// so the installation is left with nothing able to fix it and no way to ask for
// help. The updater is the only process still running at that point, which makes
// it the only place this check can live.
//
// One replica, not however many there were. Swarm's PreviousSpec is the only
// record of that, and an update that got far enough to replace it would have this
// restore zero - the very state it is here to get out of. One is enough to bring
// the dashboard back, which is where an operator can set the real number.
func (s *service) EnsureHpAppRunning(ctx context.Context) (bool, error) {
	service, err := s.GetHpAppSwarmService(ctx)
	if err != nil {
		return false, hperrors.Wrap(err)
	}
	scaled := false
	err = s.dockerManager.ServiceUpdateFunc(ctx, service.ID, service,
		func(_ int, svc *swarm.Service) (bool, error) {
			// A global app service is not a shape HivePaaS deploys, but reading
			// Replicas through a nil Replicated would panic in the one path that
			// must not.
			if svc.Spec.Mode.Replicated == nil || svc.Spec.Mode.Replicated.Replicas == nil {
				return false, nil
			}
			if *svc.Spec.Mode.Replicated.Replicas > 0 {
				return false, nil
			}
			svc.Spec.Mode.Replicated.Replicas = new(uint64(1))
			scaled = true
			return true, nil
		}, serviceUpdateRetryMax, 0)
	if err != nil {
		return false, hperrors.Wrap(err)
	}
	return scaled, nil
}
