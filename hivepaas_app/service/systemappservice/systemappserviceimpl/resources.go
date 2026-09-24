package systemappserviceimpl

import (
	"context"
	"errors"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/systemappservice"
)

// serviceUpdateRetryMax is how often a resource change is retried when the
// service moved on under it.
const serviceUpdateRetryMax = 2

func (s *service) ReadResources(ctx context.Context, app *entity.App) (*systemappservice.Resources, error) {
	if app.ServiceID == "" {
		return nil, nil
	}
	inspected, err := s.dockerManager.ServiceInspect(ctx, app.ServiceID)
	if err != nil {
		if errors.Is(err, hperrors.ErrNotFound) {
			return nil, nil
		}
		return nil, hperrors.Wrap(err)
	}
	res := systemappservice.ResourcesOf(&inspected.Service.Spec)
	return &res, nil
}

func (s *service) SetResources(ctx context.Context, app *entity.App, res systemappservice.Resources) error {
	err := s.dockerManager.ServiceUpdateFunc(ctx, app.ServiceID, nil,
		func(_ int, svc *swarm.Service) (bool, error) {
			return systemappservice.ApplyResources(&svc.Spec, res), nil
		}, serviceUpdateRetryMax, 0)
	return hperrors.Wrap(err)
}
