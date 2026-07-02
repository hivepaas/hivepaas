package hpappserviceimpl

import (
	"context"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

func (s *service) GetHpUpdaterSwarmService(ctx context.Context) (*swarm.Service, error) {
	service, err := s.dockerManager.ServiceGetByName(ctx, base.HivepaasUpdaterServiceName, false)
	if err != nil {
		return nil, apperrors.New(err)
	}
	return service, nil
}

func (s *service) RestartHpUpdaterSwarmService(ctx context.Context) error {
	service, err := s.GetHpUpdaterSwarmService(ctx)
	if err != nil {
		return apperrors.New(err)
	}

	service.Spec.TaskTemplate.ForceUpdate++
	_, err = s.dockerManager.ServiceUpdate(ctx, service.ID, &service.Version, &service.Spec)
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}

func (s *service) ShutdownHpUpdaterSwarmService(ctx context.Context) error {
	service, err := s.GetHpUpdaterSwarmService(ctx)
	if err != nil {
		return apperrors.New(err)
	}

	if service.Spec.Mode.Replicated == nil || *service.Spec.Mode.Replicated.Replicas == 0 {
		return nil
	}
	service.Spec.Mode.Replicated.Replicas = new(uint64(0))

	_, err = s.dockerManager.ServiceUpdate(ctx, service.ID, &service.Version, &service.Spec)
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}
