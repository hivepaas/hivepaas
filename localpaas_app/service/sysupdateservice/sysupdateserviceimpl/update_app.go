package sysupdateserviceimpl

import (
	"context"
	"errors"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/pkg/applog"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
)

const (
	mainAppServiceUpdateCheckInterval = time.Second * 5
	workerServiceUpdateCheckInterval  = time.Second * 5
)

func (s *service) scaleMainAppService(
	ctx context.Context,
	replicas uint64,
	data *sysUpdateData,
) error {
	mainAppSvc, err := s.lpAppService.GetLpAppSwarmService(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}
	if data.CurrentAppReplicas == nil {
		data.CurrentAppReplicas = mainAppSvc.Spec.Mode.Replicated.Replicas
	}
	if *mainAppSvc.Spec.Mode.Replicated.Replicas == replicas {
		return nil
	}

	err = s.scaleServiceReplicas(ctx, mainAppSvc, replicas)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (s *service) scaleWorkerService(
	ctx context.Context,
	replicas uint64,
	data *sysUpdateData,
) error {
	workerSvc, err := s.lpAppService.GetLpWorkerSwarmService(ctx)
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return apperrors.Wrap(err)
	}
	if workerSvc == nil {
		return nil
	}

	if data.CurrentWorkerReplicas == nil {
		data.CurrentWorkerReplicas = workerSvc.Spec.Mode.Replicated.Replicas
	}
	if *workerSvc.Spec.Mode.Replicated.Replicas == replicas {
		return nil
	}

	err = s.scaleServiceReplicas(ctx, workerSvc, replicas)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (s *service) updateMainAppService(
	ctx context.Context,
	data *sysUpdateData,
) (err error) {
	args := gofn.Must(data.Task.ArgsAsSystemUpdate())

	start := timeutil.NowUTC()
	_ = data.LogStore.Add(ctx, applog.NewOutFrame("Updating localpaas service...", applog.TsNow))
	defer func() {
		duration := timeutil.NowUTC().Sub(start)
		if err != nil {
			_ = data.LogStore.Add(ctx, applog.NewOutFrame("Updating localpaas service finished in "+
				duration.String()+" with error: "+err.Error(), applog.TsNow))
		} else {
			_ = data.LogStore.Add(ctx, applog.NewOutFrame("Updating localpaas service finished in "+
				duration.String(), applog.TsNow))
		}
	}()

	appSvc, err := s.lpAppService.GetLpAppSwarmService(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}

	appSvc.Spec.TaskTemplate.ContainerSpec.Image = args.TargetVersion.AppImage
	appSvc.Spec.Mode.Replicated.Replicas = data.CurrentAppReplicas

	_, err = s.dockerManager.ServiceUpdate(ctx, appSvc.ID, &appSvc.Version, &appSvc.Spec)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Wait for the update to finish
	appSvc, err = s.dockerManager.ServiceUpdateWait(ctx, appSvc.ID, mainAppServiceUpdateCheckInterval)
	if err != nil {
		return apperrors.Wrap(err)
	}
	if appSvc.UpdateStatus != nil && appSvc.UpdateStatus.State == swarm.UpdateStateRollbackCompleted {
		_ = data.LogStore.Add(ctx, applog.NewWarnFrame("service localpaas is rolled back",
			applog.TsNow))
		return apperrors.Wrap(apperrors.ErrActionFailed)
	}

	return nil
}

func (s *service) updateWorkerService(
	ctx context.Context,
	data *sysUpdateData,
) (err error) {
	args := gofn.Must(data.Task.ArgsAsSystemUpdate())

	start := timeutil.NowUTC()
	_ = data.LogStore.Add(ctx, applog.NewOutFrame("Updating localpaas worker service...", applog.TsNow))
	defer func() {
		duration := timeutil.NowUTC().Sub(start)
		if err != nil {
			_ = data.LogStore.Add(ctx, applog.NewOutFrame("Updating localpaas worker service finished in "+
				duration.String()+" with error: "+err.Error(), applog.TsNow))
		} else {
			_ = data.LogStore.Add(ctx, applog.NewOutFrame("Updating localpaas worker service finished in "+
				duration.String(), applog.TsNow))
		}
	}()

	workerSvc, err := s.lpAppService.GetLpWorkerSwarmService(ctx)
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return apperrors.Wrap(err)
	}
	if workerSvc == nil {
		return nil
	}

	workerSvc.Spec.TaskTemplate.ContainerSpec.Image = args.TargetVersion.AppImage
	workerSvc.Spec.Mode.Replicated.Replicas = data.CurrentWorkerReplicas

	_, err = s.dockerManager.ServiceUpdate(ctx, workerSvc.ID, &workerSvc.Version, &workerSvc.Spec)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Wait for the update to finish
	workerSvc, err = s.dockerManager.ServiceUpdateWait(ctx, workerSvc.ID, workerServiceUpdateCheckInterval)
	if err != nil {
		return apperrors.Wrap(err)
	}
	if workerSvc.UpdateStatus != nil && workerSvc.UpdateStatus.State == swarm.UpdateStateRollbackCompleted {
		_ = data.LogStore.Add(ctx, applog.NewWarnFrame("service localpaas worker is rolled back",
			applog.TsNow))
		return apperrors.Wrap(apperrors.ErrActionFailed)
	}

	return nil
}
