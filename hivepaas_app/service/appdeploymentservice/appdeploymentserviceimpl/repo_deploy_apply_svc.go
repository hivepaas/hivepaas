package appdeploymentserviceimpl

import (
	"context"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/dockerhelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/placementservice"
)

const (
	dockerServiceApplyRetryMax = 2
)

func (s *service) repoDeployStepServiceApply(
	ctx context.Context,
	db database.IDB,
	data *repoDeploymentData,
) (err error) {
	data.Step = stepServiceApply
	deployment := data.Deployment
	repoSource := deployment.Settings.RepoSource

	s.addStepStartLog(ctx, data.appDeploymentData, "Applying changes to service...")
	defer s.addStepEndLog(ctx, data.appDeploymentData, timeutil.NowUTC(), err)

	var regAuthHeader string
	if repoSource.PushToRegistry.ID != "" {
		regAuth := data.RefObjects.RefSettings[repoSource.PushToRegistry.ID]
		regAuthHeader, err = regAuth.MustAsRegistryAuth().GenerateAuthHeader()
		if err != nil {
			return apperrors.Wrap(err)
		}
	}

	queryRegistry := false
	placementReq := &placementservice.ApplyPlacementSettingsReq{
		App:                data.App,
		SkipSavingToDocker: true,
	}

	err = s.dockerManager.ServiceUpdateFunc(ctx, data.App.ServiceID,
		func(i int, svc *swarm.Service) error {
			if i > 0 {
				queryRegistry = true
			}
			contSpec := svc.Spec.TaskTemplate.ContainerSpec
			contSpec.Image = data.Deployment.Output.ImageTags[0]
			contSpec.Dir = deployment.Settings.WorkingDir
			dockerhelper.ContainerCommandApply(contSpec, deployment.Settings.Command)

			placementReq.Service = svc
			_, err := s.placementService.ApplyPlacementSettings(ctx, db, placementReq)
			if err != nil {
				return apperrors.Wrap(err)
			}
			return nil
		}, dockerServiceApplyRetryMax, 0, 0,
		func(options *client.ServiceUpdateOptions) {
			options.EncodedRegistryAuth = regAuthHeader
			options.QueryRegistry = queryRegistry
		})
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}
