package tasksystemupdate

import (
	"context"
	"time"

	"github.com/moby/moby/api/types/swarm"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/pkg/applog"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
)

const (
	redisServiceUpdateCheckInterval = time.Second * 5
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
	if redisSvc.Spec.UpdateConfig == nil {
		redisSvc.Spec.UpdateConfig = &swarm.UpdateConfig{}
	}
	redisSvc.Spec.UpdateConfig.FailureAction = swarm.UpdateFailureActionRollback
	redisSvc.Spec.UpdateConfig.MaxFailureRatio = 0.5

	_, err = e.dockerManager.ServiceUpdate(ctx, redisSvc.ID, &redisSvc.Version, &redisSvc.Spec)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Wait for the update to finish
	redisSvc, err = e.dockerManager.ServiceUpdateWait(ctx, redisSvc.ID, redisServiceUpdateCheckInterval)
	if err != nil {
		return apperrors.Wrap(err)
	}
	if redisSvc.UpdateStatus != nil && redisSvc.UpdateStatus.State == swarm.UpdateStateRollbackCompleted {
		return apperrors.New(apperrors.ErrActionFailed).WithMsgLog("service redis is rolled back")
	}

	return nil
}
