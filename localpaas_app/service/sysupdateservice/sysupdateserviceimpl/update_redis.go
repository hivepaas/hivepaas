package sysupdateserviceimpl

import (
	"context"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/pkg/applog"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
)

const (
	redisServiceUpdateCheckInterval = time.Second * 5
)

func (s *service) updateRedisService(
	ctx context.Context,
	data *sysUpdateData,
) (err error) {
	args := gofn.Must(data.Task.ArgsAsSystemUpdate())
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

	redisSvc, err := s.lpAppService.GetLpCacheSwarmService(ctx)
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

	_, err = s.dockerManager.ServiceUpdate(ctx, redisSvc.ID, &redisSvc.Version, &redisSvc.Spec)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Wait for the update to finish
	redisSvc, err = s.dockerManager.ServiceUpdateWait(ctx, redisSvc.ID, redisServiceUpdateCheckInterval)
	if err != nil {
		return apperrors.Wrap(err)
	}
	if redisSvc.UpdateStatus != nil && redisSvc.UpdateStatus.State == swarm.UpdateStateRollbackCompleted {
		_ = data.LogStore.Add(ctx, applog.NewWarnFrame("service redis is rolled back",
			applog.TsNow))
		return apperrors.Wrap(apperrors.ErrActionFailed)
	}

	return nil
}
