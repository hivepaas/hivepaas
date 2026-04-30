package tasksystemupdate

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/pkg/applog"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
)

func (e *Executor) updateTraefikService(
	ctx context.Context,
	data *taskData,
) (err error) {
	args := data.UpdateArgs
	if args.TargetVersion.TraefikImage == "" {
		return nil
	}

	start := timeutil.NowUTC()
	_ = data.LogStore.Add(ctx, applog.NewOutFrame("Updating traefik service...", applog.TsNow))
	defer func() {
		duration := timeutil.NowUTC().Sub(start)
		if err != nil {
			_ = data.LogStore.Add(ctx, applog.NewOutFrame("Updating traefik service finished in "+duration.String()+
				" with error: "+err.Error(), applog.TsNow))
		} else {
			_ = data.LogStore.Add(ctx, applog.NewOutFrame("Updating traefik service finished in "+duration.String(),
				applog.TsNow))
		}
	}()

	traefikSvc, err := e.traefikService.GetTraefikSwarmService(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}

	traefikSvc.Spec.TaskTemplate.ContainerSpec.Image = args.TargetVersion.TraefikImage
	_, err = e.dockerManager.ServiceUpdate(ctx, traefikSvc.ID, &traefikSvc.Version, &traefikSvc.Spec)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}
