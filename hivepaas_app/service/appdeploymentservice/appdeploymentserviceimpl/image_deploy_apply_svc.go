package appdeploymentserviceimpl

import (
	"context"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/placementservice"
	"github.com/hivepaas/hivepaas/services/docker/dockerhelper"
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

	err = s.dockerManager.ServiceUpdateFunc(ctx, data.App.ServiceID, nil,
		func(i int, svc *swarm.Service) (bool, error) {
			if i > 0 {
				queryRegistry = true
			}
			contSpec := svc.Spec.TaskTemplate.ContainerSpec
			contSpec.Image = imageSource.Image
			contSpec.Dir = deployment.Settings.WorkingDir
			dockerhelper.ContainerCommandApply(contSpec, deployment.Settings.Command)

			placementReq.Service = svc
			_, err := s.placementService.ApplyPlacementSettings(ctx, db, placementReq)
			if err != nil {
				return false, apperrors.Wrap(err)
			}
			return true, nil
		}, dockerServiceApplyRetryMax, 0,
		func(options *client.ServiceUpdateOptions) {
			options.EncodedRegistryAuth = data.RegAuthHeader
			options.QueryRegistry = queryRegistry
		})
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Save the used image in the output
	data.Deployment.Output.ImageTags = append(data.Deployment.Output.ImageTags, imageSource.Image)

	return nil
}
