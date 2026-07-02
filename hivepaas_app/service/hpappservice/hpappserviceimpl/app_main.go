package hpappserviceimpl

import (
	"context"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

func (s *service) GetHpAppSwarmService(ctx context.Context) (*swarm.Service, error) {
	service, err := s.dockerManager.ServiceGetByName(ctx, base.HivepaasAppServiceName, false)
	if err != nil {
		return nil, apperrors.New(err)
	}
	return service, nil
}

func (s *service) RestartHpAppSwarmService(ctx context.Context) error {
	service, err := s.GetHpAppSwarmService(ctx)
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

func (s *service) GetHpAppTasks(ctx context.Context) ([]swarm.Task, error) {
	service, err := s.GetHpAppSwarmService(ctx)
	if err != nil {
		return nil, apperrors.New(err)
	}

	resp, err := s.dockerManager.ServiceTaskList(ctx, service.ID, nil)
	if err != nil {
		return nil, apperrors.New(err)
	}
	return resp.Items, nil
}
