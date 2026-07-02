package hpappserviceimpl

import (
	"context"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

func (s *service) GetHpWorkerSwarmService(ctx context.Context) (*swarm.Service, error) {
	service, err := s.dockerManager.ServiceGetByName(ctx, base.HivepaasWorkerServiceName, false)
	if err != nil {
		return nil, apperrors.New(err)
	}
	return service, nil
}

func (s *service) RestartHpWorkerSwarmService(ctx context.Context) error {
	service, err := s.GetHpWorkerSwarmService(ctx)
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

func (s *service) SyncHpWorkerSwarmServiceConfig(
	mainAppSvc, workerSvc *swarm.Service,
) {
	workerSvc.Spec.TaskTemplate.ContainerSpec.Image = mainAppSvc.Spec.TaskTemplate.ContainerSpec.Image
	workerSvc.Spec.TaskTemplate.ContainerSpec.Command = mainAppSvc.Spec.TaskTemplate.ContainerSpec.Command
	workerSvc.Spec.TaskTemplate.ContainerSpec.Args = mainAppSvc.Spec.TaskTemplate.ContainerSpec.Args

	// TODO: sync Envs

	// Make sure the worker service has the same storages as the main service
	workerSvc.Spec.TaskTemplate.ContainerSpec.Mounts = mainAppSvc.Spec.TaskTemplate.ContainerSpec.Mounts
}
