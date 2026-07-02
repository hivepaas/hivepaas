package sysupdateserviceimpl

import (
	"context"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
)

func (s *service) scaleServiceReplicas(
	ctx context.Context,
	service *swarm.Service,
	replicas uint64,
) error {
	if service.Spec.Mode.Replicated == nil {
		return nil
	}
	if *service.Spec.Mode.Replicated.Replicas == replicas {
		return nil
	}
	service.Spec.Mode.Replicated.Replicas = &replicas
	_, err := s.dockerManager.ServiceUpdate(ctx, service.ID, &service.Version, &service.Spec)
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}
