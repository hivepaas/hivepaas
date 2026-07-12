package hpappserviceimpl

import (
	"context"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

func (s *service) GetHpDbSwarmService(ctx context.Context) (*swarm.Service, error) {
	service, err := s.dockerManager.ServiceGetByName(ctx, base.HivepaasDbServiceName, false)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	return service, nil
}

func (s *service) RestartHpDbSwarmService(ctx context.Context) error {
	service, err := s.GetHpDbSwarmService(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}

	service.Spec.TaskTemplate.ForceUpdate++
	_, err = s.dockerManager.ServiceUpdate(ctx, service.ID, &service.Version, &service.Spec)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}
