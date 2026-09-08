package hpappserviceimpl

import (
	"context"
	"strings"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func (s *service) GetHpWorkerSwarmService(ctx context.Context) (*swarm.Service, error) {
	service, err := s.dockerManager.ServiceGetByName(ctx, base.HivepaasWorkerServiceName, false)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return service, nil
}

func (s *service) RestartHpWorkerSwarmService(ctx context.Context) error {
	service, err := s.GetHpWorkerSwarmService(ctx)
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

func (s *service) SyncHpWorkerSwarmServiceConfig(
	mainAppSvc, workerSvc *swarm.Service,
) {
	if mainAppSvc == nil || workerSvc == nil {
		return
	}
	if mainAppSvc.Spec.TaskTemplate.ContainerSpec == nil {
		mainAppSvc.Spec.TaskTemplate.ContainerSpec = &swarm.ContainerSpec{}
	}
	if workerSvc.Spec.TaskTemplate.ContainerSpec == nil {
		workerSvc.Spec.TaskTemplate.ContainerSpec = &swarm.ContainerSpec{}
	}

	mainContainer := mainAppSvc.Spec.TaskTemplate.ContainerSpec
	workerContainer := workerSvc.Spec.TaskTemplate.ContainerSpec

	workerContainer.Image = mainContainer.Image
	workerContainer.Command = mainContainer.Command
	workerContainer.Args = mainContainer.Args
	workerContainer.Env = syncWorkerEnvs(mainContainer.Env)

	// Make sure the worker service has the same storages as the main service
	workerContainer.Mounts = mainContainer.Mounts
}

const (
	envHpRunMode       = "HP_RUN_MODE"
	envHpRunModeWorker = "HP_RUN_MODE=worker"
)

func syncWorkerEnvs(mainEnvs []string) []string {
	workerEnvs := make([]string, 0, len(mainEnvs)+1)
	runModeSet := false
	for _, env := range mainEnvs {
		k, _, _ := strings.Cut(env, "=")
		if k == envHpRunMode {
			workerEnvs = append(workerEnvs, envHpRunModeWorker)
			runModeSet = true
		} else {
			workerEnvs = append(workerEnvs, env)
		}
	}
	if !runModeSet {
		workerEnvs = append(workerEnvs, envHpRunModeWorker)
	}
	return workerEnvs
}
