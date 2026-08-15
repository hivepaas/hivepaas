package appdeploymentserviceimpl

import (
	"context"
	"errors"
	"time"

	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/dockerhelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/placementservice"
)

func (s *service) imageDeployStepServiceApply(
	ctx context.Context,
	db database.IDB,
	data *imageDeploymentData,
) (err error) {
	data.Step = stepServiceApply
	deployment := data.Deployment
	imageSource := deployment.Settings.ImageSource

	s.addStepStartLog(ctx, data.appDeploymentData, "Applying changes to service...")
	defer s.addStepEndLog(ctx, data.appDeploymentData, timeutil.NowUTC(), err)

	queryRegistry := false
	placementReq := &placementservice.ApplyPlacementSettingsReq{
		App:                data.App,
		SkipSavingToDocker: true,
	}

	for i := range dockerServiceApplyRetryMax + 1 {
		if i > 0 {
			queryRegistry = true
			select {
			case <-ctx.Done():
				err = apperrors.Wrap(ctx.Err())
				break
			case <-time.After(time.Duration(1+i) * time.Second):
			}
		}

		inspect, e := s.dockerManager.ServiceInspect(ctx, data.App.ServiceID)
		if e != nil { // error, need to retry
			if errors.Is(e, apperrors.ErrNotFound) {
				err = apperrors.Wrap(e)
				break
			}
			err = apperrors.Wrap(e)
			continue
		}

		service := &inspect.Service
		spec := &service.Spec
		contSpec := spec.TaskTemplate.ContainerSpec
		contSpec.Image = imageSource.Image
		contSpec.Dir = deployment.Settings.WorkingDir
		dockerhelper.ContainerCommandApply(contSpec, deployment.Settings.Command)

		// Apply placement settings
		placementReq.Service = service
		_, e = s.placementService.ApplyPlacementSettings(ctx, db, placementReq)
		if e != nil { // error, need to retry
			err = apperrors.Wrap(e)
			continue
		}

		_, e = s.dockerManager.ServiceUpdate(ctx, data.App.ServiceID, &service.Version, spec,
			func(options *client.ServiceUpdateOptions) {
				options.EncodedRegistryAuth = data.RegAuthHeader
				options.QueryRegistry = queryRegistry
			})
		if e != nil { // error, need to retry
			err = apperrors.Wrap(e)
			continue
		}
		// successful, no need to retry
		err = nil
		break
	}
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Save the used image in the output
	data.Deployment.Output.ImageTags = append(data.Deployment.Output.ImageTags, imageSource.Image)

	return nil
}
