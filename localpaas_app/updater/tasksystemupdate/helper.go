package tasksystemupdate

import (
	"context"

	"github.com/moby/moby/api/types/swarm"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
)

func (e *Executor) scaleServiceReplicas(
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
	_, err := e.dockerManager.ServiceUpdate(ctx, service.ID, &service.Version, &service.Spec)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}
