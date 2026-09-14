package sysupdateserviceimpl

import (
	"context"
	"errors"

	"github.com/moby/moby/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func (s *service) scaleMainAppService(
	ctx context.Context,
	replicas uint64,
	data *sysUpdateData,
) error {
	mainAppSvc, err := s.hpAppService.GetHpAppSwarmService(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	// Remembered before anything is changed: this is what onAfterSystemUpdate
	// scales back to when the update ends, however it ends.
	if data.CurrentAppReplicas == nil && mainAppSvc.Spec.Mode.Replicated != nil {
		data.CurrentAppReplicas = mainAppSvc.Spec.Mode.Replicated.Replicas
	}

	err = s.scaleServiceReplicas(ctx, mainAppSvc, replicas)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (s *service) scaleWorkerService(
	ctx context.Context,
	replicas uint64,
	data *sysUpdateData,
) error {
	workerSvc, err := s.getWorkerSwarmService(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if workerSvc == nil {
		return nil
	}

	if data.CurrentWorkerReplicas == nil && workerSvc.Spec.Mode.Replicated != nil {
		data.CurrentWorkerReplicas = workerSvc.Spec.Mode.Replicated.Replicas
	}

	err = s.scaleServiceReplicas(ctx, workerSvc, replicas)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

// getWorkerSwarmService reports a missing worker as no service rather than as an
// error: an installation can run the worker inside the main app instead, and
// then there is no second service to update or to scale.
func (s *service) getWorkerSwarmService(ctx context.Context) (*swarm.Service, error) {
	svc, err := s.hpAppService.GetHpWorkerSwarmService(ctx)
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return nil, hperrors.Wrap(err)
	}
	return svc, nil
}

func (s *service) updateMainAppService(
	ctx context.Context,
	data *sysUpdateData,
) error {
	args := gofn.Must(data.Task.ArgsAsSystemUpdate())

	err := s.updateServiceImage(ctx, data, serviceImageUpdate{
		What:        "hivepaas",
		Component:   base.HivepaasAppKey,
		TargetImage: args.TargetVersion.AppImage,
		Fetch: func(ctx context.Context) (*swarm.Service, error) {
			return s.hpAppService.GetHpAppSwarmService(ctx)
		},
		// stopServices scaled this to zero before any image moved; this is where
		// it comes back. A step that decides the image is already current leaves
		// that to onAfterSystemUpdate, which scales both services back either way.
		Mutate: func(spec *swarm.ServiceSpec) {
			spec.Mode.Replicated.Replicas = data.CurrentAppReplicas
		},
	})
	return hperrors.Wrap(err)
}

func (s *service) updateWorkerService(
	ctx context.Context,
	data *sysUpdateData,
) error {
	args := gofn.Must(data.Task.ArgsAsSystemUpdate())

	err := s.updateServiceImage(ctx, data, serviceImageUpdate{
		What:        "hivepaas worker",
		Component:   base.HivepaasWorkerKey,
		TargetImage: args.TargetVersion.AppImage,
		Fetch:       s.getWorkerSwarmService,
		Mutate: func(spec *swarm.ServiceSpec) {
			spec.Mode.Replicated.Replicas = data.CurrentWorkerReplicas
		},
	})
	return hperrors.Wrap(err)
}
