package hpappserviceimpl

import (
	"context"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/gocache"
)

const (
	cacheKeyHPAgentImage = "hivepaas:agent:image"
)

func (s *service) GetHpAgentSwarmService(ctx context.Context) (*swarm.Service, error) {
	service, err := s.dockerManager.ServiceGetByName(ctx, base.HivepaasAgentServiceName, false)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	return service, nil
}

func (s *service) RestartHpAgentSwarmService(ctx context.Context) error {
	service, err := s.GetHpAgentSwarmService(ctx)
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

func (s *service) GetHpAgentImage(ctx context.Context) string {
	image, _ := gocache.Global.GetStr(cacheKeyHPAgentImage)
	if image != "" {
		return image
	}
	if agentSvc, err := s.GetHpAgentSwarmService(ctx); err == nil && agentSvc != nil {
		if agentSvc.Spec.TaskTemplate.ContainerSpec != nil && agentSvc.Spec.TaskTemplate.ContainerSpec.Image != "" {
			image = agentSvc.Spec.TaskTemplate.ContainerSpec.Image
			_ = gocache.Global.Set(cacheKeyHPAgentImage, image, 0)
			return image
		}
	}
	return ""
}
