package lpappserviceimpl

import (
	"context"

	"github.com/moby/moby/api/types/swarm"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
)

func (s *service) GetLpAppSwarmService(ctx context.Context) (*swarm.Service, error) {
	service, err := s.dockerManager.ServiceGetByName(ctx, base.LocalpaasAppServiceName, false)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	return service, nil
}

func (s *service) RestartLpAppSwarmService(ctx context.Context) error {
	service, err := s.GetLpAppSwarmService(ctx)
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

func (s *service) GetLpAppTasks(ctx context.Context) ([]swarm.Task, error) {
	service, err := s.GetLpAppSwarmService(ctx)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp, err := s.dockerManager.ServiceTaskList(ctx, service.ID, "")
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	return resp.Items, nil
}
