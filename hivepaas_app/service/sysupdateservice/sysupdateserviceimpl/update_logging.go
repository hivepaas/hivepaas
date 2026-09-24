package sysupdateserviceimpl

import (
	"context"
	"errors"

	"github.com/moby/moby/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/imageref"
)

// updateLoggingService moves the logging stack to the images this release names.
//
// It is two apps, not one: the backend stores and answers queries, the
// collector ships lines to it from every node. They are separate upstream
// repositories with separate tags, so the release names them separately too.
//
// Neither is in the stack file. They are apps HivePaaS provisions in its hidden
// project when logging is switched on and removes when it is switched off or
// handed to a system somebody else runs - so an app that is not there is the
// ordinary case, handled the way the worker service already is, and not a
// reason to fail the update.
func (s *service) updateLoggingService(
	ctx context.Context,
	db database.IDB,
	data *sysUpdateData,
) error {
	args := gofn.Must(data.Task.ArgsAsSystemUpdate())

	// The backend first, so the collector is never the only one on the new
	// version writing into an older store.
	steps := []systemAppImageUpdate{
		{What: "victoria-logs", Key: base.HivepaasVictoriaLogsKey, TargetImage: args.TargetVersion.VictoriaLogsImage},
		{What: "vlagent", Key: base.HivepaasVlagentKey, TargetImage: args.TargetVersion.VlagentImage},
	}
	for _, step := range steps {
		if err := s.updateSystemAppImage(ctx, db, data, step); err != nil {
			return hperrors.Wrap(err)
		}
	}
	return nil
}

// systemAppImageUpdate is one app HivePaaS provisioned for itself, and the image
// the release moves it to.
type systemAppImageUpdate struct {
	What        string
	Key         string
	TargetImage string
}

// updateSystemAppImage moves the app's service the way every other service of
// the update moves - with the version check and swarm's rollback armed - and
// then records the image in the app's deployment settings. The deployment
// settings are what a deployment applies: without the second half, the app's
// next deployment would put the old image back.
func (s *service) updateSystemAppImage(
	ctx context.Context,
	db database.IDB,
	data *sysUpdateData,
	step systemAppImageUpdate,
) error {
	if step.TargetImage == "" {
		return nil
	}
	app, err := s.systemAppService.LoadApp(ctx, db, step.Key)
	if err != nil {
		return hperrors.Wrap(err)
	}

	err = s.updateServiceImage(ctx, data, serviceImageUpdate{
		What:        step.What,
		Component:   step.Key,
		TargetImage: step.TargetImage,
		Fetch: func(ctx context.Context) (*swarm.Service, error) {
			return s.systemAppSwarmService(ctx, app)
		},
	})
	if err != nil || app == nil {
		return hperrors.Wrap(err)
	}

	// Recorded only when the service runs the target now. The version check may
	// have kept what was there - a newer image, one the release does not know -
	// and the settings must say what runs, not what was asked for.
	svc, err := s.systemAppSwarmService(ctx, app)
	if err != nil || svc == nil {
		return hperrors.Wrap(err)
	}
	running := imageref.Parse(svc.Spec.TaskTemplate.ContainerSpec.Image)
	target := imageref.Parse(step.TargetImage)
	if !imageref.SameRepository(running.Repository, target.Repository) || running.Tag != target.Tag {
		return nil
	}
	return hperrors.Wrap(s.systemAppService.RecordImage(ctx, db, app, step.TargetImage))
}

// systemAppSwarmService is the app's service, or nil when the app or its
// service is not there - logging switched off, or its service removed by hand.
func (s *service) systemAppSwarmService(ctx context.Context, app *entity.App) (*swarm.Service, error) {
	if app == nil || app.ServiceID == "" {
		return nil, nil
	}
	inspected, err := s.dockerManager.ServiceInspect(ctx, app.ServiceID)
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
