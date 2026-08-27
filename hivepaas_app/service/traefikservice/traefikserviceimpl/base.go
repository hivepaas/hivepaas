package traefikserviceimpl

import (
	"context"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func (s *service) GetTraefikSwarmService(ctx context.Context) (*swarm.Service, error) {
	if s.dockerManager == nil {
		return nil, nil
	}
	service, err := s.dockerManager.ServiceGetByName(ctx, base.HivepaasTraefikServiceName, false)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return service, nil
}

func (s *service) RestartTraefikSwarmService(ctx context.Context) error {
	service, err := s.GetTraefikSwarmService(ctx)
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
