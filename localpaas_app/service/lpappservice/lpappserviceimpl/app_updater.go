package lpappserviceimpl

import (
	"context"

	"github.com/moby/moby/api/types/swarm"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
)

func (s *service) GetLpUpdaterSwarmService(ctx context.Context) (*swarm.Service, error) {
	service, err := s.dockerManager.ServiceGetByName(ctx, base.LocalpaasUpdaterServiceName, false)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	return service, nil
}

func (s *service) RestartLpUpdaterSwarmService(ctx context.Context) error {
	service, err := s.GetLpUpdaterSwarmService(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}

	err = s.dockerManager.ServiceForceUpdate(ctx, service.ID)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (s *service) ShutdownLpUpdaterSwarmService(ctx context.Context) error {
	service, err := s.GetLpUpdaterSwarmService(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}

	if service.Spec.Mode.Replicated == nil || *service.Spec.Mode.Replicated.Replicas == 0 {
		return nil
	}
	service.Spec.Mode.Replicated.Replicas = new(uint64(0))

	_, err = s.dockerManager.ServiceUpdate(ctx, service.ID, &service.Version, &service.Spec)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}
