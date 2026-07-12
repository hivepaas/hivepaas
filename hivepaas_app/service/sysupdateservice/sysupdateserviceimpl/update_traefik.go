package sysupdateserviceimpl

import (
	"context"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

const (
	traefikServiceUpdateCheckInterval = time.Second * 5
)

func (s *service) updateTraefikService(
	ctx context.Context,
	data *sysUpdateData,
) (err error) {
	args := gofn.Must(data.Task.ArgsAsSystemUpdate())
	if args.TargetVersion.TraefikImage == "" {
		return nil
	}

	start := timeutil.NowUTC()
	_ = data.LogStore.Add(ctx, tasklog.NewOutFrame("Updating traefik service...", tasklog.TsNow))
	defer func() {
		duration := timeutil.NowUTC().Sub(start)
		if err != nil {
			_ = data.LogStore.Add(ctx, tasklog.NewOutFrame("Updating traefik service finished in "+duration.String()+
				" with error: "+err.Error(), tasklog.TsNow))
		} else {
			_ = data.LogStore.Add(ctx, tasklog.NewOutFrame("Updating traefik service finished in "+duration.String(),
				tasklog.TsNow))
		}
	}()

	traefikSvc, err := s.traefikService.GetTraefikSwarmService(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}

	traefikSvc.Spec.TaskTemplate.ContainerSpec.Image = args.TargetVersion.TraefikImage
	if traefikSvc.Spec.UpdateConfig == nil {
		traefikSvc.Spec.UpdateConfig = &swarm.UpdateConfig{}
	}
	traefikSvc.Spec.UpdateConfig.FailureAction = swarm.UpdateFailureActionRollback
	traefikSvc.Spec.UpdateConfig.MaxFailureRatio = 0.5

	_, err = s.dockerManager.ServiceUpdate(ctx, traefikSvc.ID, &traefikSvc.Version, &traefikSvc.Spec)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Wait for the update to finish
	traefikSvc, err = s.dockerManager.ServiceUpdateWait(ctx, traefikSvc.ID, traefikServiceUpdateCheckInterval)
	if err != nil {
		return apperrors.Wrap(err)
	}
	if traefikSvc.UpdateStatus != nil && traefikSvc.UpdateStatus.State == swarm.UpdateStateRollbackCompleted {
		_ = data.LogStore.Add(ctx, tasklog.NewWarnFrame("service traefik is rolled back",
			tasklog.TsNow))
		return apperrors.Wrap(apperrors.ErrActionFailed)
	}

	return nil
}
