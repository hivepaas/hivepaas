package sysupdateserviceimpl

import (
	"context"

	"github.com/moby/moby/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func (s *service) updateTraefikService(
	ctx context.Context,
	data *sysUpdateData,
) error {
	args := gofn.Must(data.Task.ArgsAsSystemUpdate())

	err := s.updateServiceImage(ctx, data, serviceImageUpdate{
		What:        "traefik",
		Component:   base.HivepaasTraefikKey,
		TargetImage: args.TargetVersion.TraefikImage,
		Fetch: func(ctx context.Context) (*swarm.Service, error) {
			return s.traefikService.GetTraefikSwarmService(ctx)
		},
	})
	return hperrors.Wrap(err)
}
