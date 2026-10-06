package sysupdateserviceimpl

import (
	"context"

	"github.com/moby/moby/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/traefikservice"
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
		Align: alignTraefik,
	})
	return hperrors.Wrap(err)
}

// alignTraefik brings traefik to what this release writes, its image moving or
// not: its lines get its identity, for its access log to be counted from, and
// the access log is written as the release writes it. It reports whether it
// changed the spec.
func alignTraefik(spec *swarm.ServiceSpec) bool {
	marked := traefikservice.WithAccessLogIdentity(spec)
	written := traefikservice.WithAccessLogArgs(spec)
	return marked || written
}
