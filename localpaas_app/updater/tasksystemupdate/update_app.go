package tasksystemupdate

import (
	"context"
	"errors"
	"time"

	"github.com/moby/moby/api/types/swarm"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/pkg/applog"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
)

const (
	mainAppServiceUpdateCheckInterval = time.Second * 5
	workerServiceUpdateCheckInterval  = time.Second * 5
)

func (e *Executor) scaleMainAppService(
	ctx context.Context,
	replicas uint64,
	data *taskData,
) error {
	mainAppSvc, err := e.lpAppService.GetLpAppSwarmService(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}
	if data.CurrentAppReplicas == nil {
		data.CurrentAppReplicas = mainAppSvc.Spec.Mode.Replicated.Replicas
	}
	if *mainAppSvc.Spec.Mode.Replicated.Replicas == replicas {
		return nil
	}

	err = e.scaleServiceReplicas(ctx, mainAppSvc, replicas)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (e *Executor) scaleWorkerService(
	ctx context.Context,
	replicas uint64,
	data *taskData,
) error {
	workerSvc, err := e.lpAppService.GetLpWorkerSwarmService(ctx)
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

	err = e.scaleServiceReplicas(ctx, workerSvc, replicas)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (e *Executor) updateMainAppService(
	ctx context.Context,
	data *taskData,
) (err error) {
	args := data.UpdateArgs

	start := timeutil.NowUTC()
	_ = data.LogStore.Add(ctx, applog.NewOutFrame("Updating localpaas service...", applog.TsNow))
	defer func() {
		duration := timeutil.NowUTC().Sub(start)
		if err != nil {
			_ = data.LogStore.Add(ctx, applog.NewOutFrame("Updating localpaas service finished in "+duration.String()+
				" with error: "+err.Error(), applog.TsNow))
		} else {
			_ = data.LogStore.Add(ctx, applog.NewOutFrame("Updating localpaas service finished in "+duration.String(),
				applog.TsNow))
		}
	}()

	appSvc, err := e.lpAppService.GetLpAppSwarmService(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}

	appSvc.Spec.TaskTemplate.ContainerSpec.Image = args.TargetVersion.AppImage
	appSvc.Spec.Mode.Replicated.Replicas = data.CurrentAppReplicas

	_, err = e.dockerManager.ServiceUpdate(ctx, appSvc.ID, &appSvc.Version, &appSvc.Spec)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Wait for the update to finish
	appSvc, err = e.dockerManager.ServiceUpdateWait(ctx, appSvc.ID, mainAppServiceUpdateCheckInterval)
	if err != nil {
		return apperrors.Wrap(err)
	}
	if appSvc.UpdateStatus != nil && appSvc.UpdateStatus.State == swarm.UpdateStateRollbackCompleted {
		return apperrors.New(apperrors.ErrActionFailed).WithMsgLog("service localpaas is rolled back")
	}

	return nil
}

func (e *Executor) updateWorkerService(
	ctx context.Context,
	data *taskData,
) (err error) {
	args := data.UpdateArgs

	start := timeutil.NowUTC()
	_ = data.LogStore.Add(ctx, applog.NewOutFrame("Updating worker service...", applog.TsNow))
	defer func() {
		duration := timeutil.NowUTC().Sub(start)
		if err != nil {
			_ = data.LogStore.Add(ctx, applog.NewOutFrame("Updating worker service finished in "+duration.String()+
				" with error: "+err.Error(), applog.TsNow))
		} else {
			_ = data.LogStore.Add(ctx, applog.NewOutFrame("Updating worker service finished in "+duration.String(),
				applog.TsNow))
		}
	}()

	workerSvc, err := e.lpAppService.GetLpWorkerSwarmService(ctx)
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return apperrors.Wrap(err)
	}
	if workerSvc == nil {
		return nil
	}

	workerSvc.Spec.TaskTemplate.ContainerSpec.Image = args.TargetVersion.AppImage
	workerSvc.Spec.Mode.Replicated.Replicas = data.CurrentWorkerReplicas

	_, err = e.dockerManager.ServiceUpdate(ctx, workerSvc.ID, &workerSvc.Version, &workerSvc.Spec)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Wait for the update to finish
	workerSvc, err = e.dockerManager.ServiceUpdateWait(ctx, workerSvc.ID, workerServiceUpdateCheckInterval)
	if err != nil {
		return apperrors.Wrap(err)
	}
	if workerSvc.UpdateStatus != nil && workerSvc.UpdateStatus.State == swarm.UpdateStateRollbackCompleted {
		return apperrors.New(apperrors.ErrActionFailed).WithMsgLog("service localpaas worker is rolled back")
	}

	return nil
}
