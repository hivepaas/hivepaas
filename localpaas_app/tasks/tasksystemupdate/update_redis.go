package tasksystemupdate

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/pkg/applog"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
)

func (e *Executor) updateRedisService(
	ctx context.Context,
	data *taskData,
) (err error) {
	args := data.UpdateArgs
	if args.TargetVersion.RedisImage == "" {
		return nil
	}

	start := timeutil.NowUTC()
	_ = data.LogStore.Add(ctx, applog.NewOutFrame("Updating redis service...", applog.TsNow))
	defer func() {
		duration := timeutil.NowUTC().Sub(start)
		if err != nil {
			_ = data.LogStore.Add(ctx, applog.NewOutFrame("Updating redis service finished in "+duration.String()+
				" with error: "+err.Error(), applog.TsNow))
		} else {
			_ = data.LogStore.Add(ctx, applog.NewOutFrame("Updating redis service finished in "+duration.String(),
				applog.TsNow))
		}
	}()

	redisSvc, err := e.lpAppService.GetLpCacheSwarmService(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}

	redisSvc.Spec.TaskTemplate.ContainerSpec.Image = args.TargetVersion.RedisImage
	redisSvc.Spec.Mode.Replicated.Replicas = new(uint64(1))

	_, err = e.dockerManager.ServiceUpdate(ctx, redisSvc.ID, &redisSvc.Version, &redisSvc.Spec)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}
