package lpappserviceimpl

import (
	"context"

	"github.com/docker/docker/api/types/swarm"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
)

func (s *service) GetLpUpdaterSwarmService(ctx context.Context) (*swarm.Service, error) {
	service, err := s.dockerManager.ServiceGetByName(ctx, base.LocalpaasUpdaterServiceName)
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
