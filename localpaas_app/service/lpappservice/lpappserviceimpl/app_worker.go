package lpappserviceimpl

import (
	"context"

	"github.com/moby/moby/api/types/swarm"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
)

func (s *service) GetLpWorkerSwarmService(ctx context.Context) (*swarm.Service, error) {
	service, err := s.dockerManager.ServiceGetByName(ctx, base.LocalpaasWorkerServiceName, false)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	return service, nil
}

func (s *service) RestartLpWorkerSwarmService(ctx context.Context) error {
	service, err := s.GetLpWorkerSwarmService(ctx)
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

func (s *service) SyncLpWorkerSwarmServiceConfig(
	mainAppSvc, workerSvc *swarm.Service,
) {
	workerSvc.Spec.TaskTemplate.ContainerSpec.Image = mainAppSvc.Spec.TaskTemplate.ContainerSpec.Image
	workerSvc.Spec.TaskTemplate.ContainerSpec.Command = mainAppSvc.Spec.TaskTemplate.ContainerSpec.Command
	workerSvc.Spec.TaskTemplate.ContainerSpec.Args = mainAppSvc.Spec.TaskTemplate.ContainerSpec.Args

	// TODO: sync Envs

	// Make sure the worker service has the same storages as the main service
	workerSvc.Spec.TaskTemplate.ContainerSpec.Mounts = mainAppSvc.Spec.TaskTemplate.ContainerSpec.Mounts
}
