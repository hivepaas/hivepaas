package sysupdateserviceimpl

import (
	"context"
	"errors"

	"github.com/moby/moby/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// updateLoggingService moves the logging stack to the images this release names.
//
// It is two services, not one: the backend stores and answers queries, the
// collector ships lines to it from every node. They are separate upstream
// repositories with separate tags, so the release names them separately too.
//
// Neither is in the stack file. The app creates them when logging is switched on
// and removes them when it is switched off or pointed at a backend somebody else
// runs - so a service that is not there is the ordinary case, handled the way the
// worker service already is, and not a reason to fail the update.
//
// The images also have to be what the logging service itself would deploy, or
// this would be undone: Apply rebuilds the whole swarm spec from the release, so
// an image written only here would go back to the release's on the next settings
// save. That is why loggingserviceimpl reads them from base.ReleaseInfo too.
func (s *service) updateLoggingService(
	ctx context.Context,
	data *sysUpdateData,
) error {
	args := gofn.Must(data.Task.ArgsAsSystemUpdate())

	// The backend first, so the collector is never the only one on the new
	// version writing into an older store.
	steps := []serviceImageUpdate{
		{
			What:        "victoria-logs",
			Component:   base.HivepaasVictoriaLogsKey,
			TargetImage: args.TargetVersion.VictoriaLogsImage,
			Fetch:       s.loggingSwarmServiceFunc(base.HivepaasVictoriaLogsServiceName),
		},
		{
			What:        "vlagent",
			Component:   base.HivepaasVlagentKey,
			TargetImage: args.TargetVersion.VlagentImage,
			Fetch:       s.loggingSwarmServiceFunc(base.HivepaasVlagentServiceName),
		},
	}

	for _, step := range steps {
		if err := s.updateServiceImage(ctx, data, step); err != nil {
			return hperrors.Wrap(err)
		}
	}
	return nil
}

// loggingSwarmServiceFunc looks a logging service up by name, reporting one that
// does not exist as logging being switched off rather than as an error.
func (s *service) loggingSwarmServiceFunc(name string) func(context.Context) (*swarm.Service, error) {
	return func(ctx context.Context) (*swarm.Service, error) {
		inspected, err := s.dockerManager.ServiceInspect(ctx, name)
		if err != nil {
			if errors.Is(err, hperrors.ErrNotFound) {
				return nil, nil
			}
			return nil, hperrors.Wrap(err)
		}
		if inspected == nil {
			return nil, nil
		}
		return &inspected.Service, nil
	}
}
