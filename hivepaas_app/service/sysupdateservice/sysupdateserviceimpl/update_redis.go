package sysupdateserviceimpl

import (
	"context"

	"github.com/moby/moby/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func (s *service) updateRedisService(
	ctx context.Context,
	data *sysUpdateData,
) error {
	args := gofn.Must(data.Task.ArgsAsSystemUpdate())

	err := s.updateServiceImage(ctx, data, serviceImageUpdate{
		What:        "redis",
		Component:   base.HivepaasCacheKey,
		TargetImage: args.TargetVersion.RedisImage,
		Fetch: func(ctx context.Context) (*swarm.Service, error) {
			return s.hpAppService.GetHpCacheSwarmService(ctx)
		},
		// The cache is not clustered: one replica is the only shape it has, and
		// an update is a reasonable moment to put it back if something scaled it.
		Mutate: func(spec *swarm.ServiceSpec) {
			spec.Mode.Replicated.Replicas = new(uint64(1))
		},
	})
	return hperrors.Wrap(err)
}
