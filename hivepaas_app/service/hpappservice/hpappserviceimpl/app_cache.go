package hpappserviceimpl

import (
	"context"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

func (s *service) GetHpCacheSwarmService(ctx context.Context) (*swarm.Service, error) {
	service, err := s.dockerManager.ServiceGetByName(ctx, base.HivepaasCacheServiceName, false)
	if err != nil {
		return nil, apperrors.New(err)
	}
	return service, nil
}

func (s *service) RestartHpCacheSwarmService(ctx context.Context) error {
	service, err := s.GetHpCacheSwarmService(ctx)
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
