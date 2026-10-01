package appdeploymentserviceimpl

import (
	"context"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/placementservice"
	"github.com/hivepaas/hivepaas/services/docker/dockerhelper"
)

// functionStopGraceMargin is how much longer than one call's timeout a stopping
// function is given: the runtime lets the running calls finish, for at most one
// timeout, and needs a moment to exit after them.
const functionStopGraceMargin = 10 * time.Second

func (s *service) functionDeployStepServiceApply(
	ctx context.Context,
	db database.IDB,
	data *repoDeploymentData,
) (err error) {
	data.Step = stepServiceApply
	source := data.Deployment.Settings.FunctionSource

	s.addStepStartLog(ctx, data.appDeploymentData, "Applying changes to service...")
	defer s.addStepEndLog(ctx, data.appDeploymentData, timeutil.NowUTC(), err)

	var regAuthHeader string
	if source.PushToRegistry.ID != "" {
		regAuth := data.RefObjects.RefSettings[source.PushToRegistry.ID]
		if regAuth == nil {
			return hperrors.NewMissing("Registry auth to pull image")
		}
		regAuthHeader, err = regAuth.MustAsRegistryAuth().GenerateAuthHeader()
		if err != nil {
			return hperrors.Wrap(err)
		}
	}

	queryRegistry := false
	placementReq := &placementservice.ApplyPlacementSettingsReq{
		App:                data.App,
		SkipSavingToDocker: true,
	}

	if err = s.prepareDockerAPI(ctx, db, data.appDeploymentData); err != nil {
		return hperrors.Wrap(err)
	}

	err = s.dockerManager.ServiceUpdateFunc(ctx, data.App.ServiceID, nil,
		func(i int, svc *swarm.Service) (bool, error) {
			if i > 0 {
				queryRegistry = true
			}
			contSpec := svc.Spec.TaskTemplate.ContainerSpec
			contSpec.Image = data.Deployment.Output.ImageTags[0]
			applyFunctionContainer(contSpec, source)
			s.applyContainerInit(ctx, data.appDeploymentData, contSpec)

			// The socket and the network follow the app's access on every
			// deployment, whatever a screen or an older release left on the service.
			if err := s.dockerAPIService.ApplyToService(ctx, db, data.App.ID, &svc.Spec); err != nil {
				return false, hperrors.Wrap(err)
			}
			// Mounted settings too: a refresh that failed is put right here.
			if err := s.settingMountService.ApplyToService(ctx, db, data.App, &svc.Spec); err != nil {
				return false, hperrors.Wrap(err)
			}

			placementReq.Service = svc
			if _, err := s.placementService.ApplyPlacementSettings(ctx, db, placementReq); err != nil {
				return false, hperrors.Wrap(err)
			}
			return true, nil
		}, dockerServiceApplyRetryMax, 0,
		func(options *client.ServiceUpdateOptions) {
			options.EncodedRegistryAuth = regAuthHeader
			options.QueryRegistry = queryRegistry
		})
	if err != nil {
		return hperrors.Wrap(err)
	}
	// What the update replaced goes once nothing references it; what is still
	// held is left to the next sweep rather than failing the deployment.
	_ = s.settingMountService.Sweep(ctx, data.App)

	return nil
}

// applyFunctionContainer gives a function's container what is fixed for a
// function: the runtime's command, working directory and health check, whatever
// a deployment setting or an earlier image left; and a stop grace period that
// lets the calls running at a stop finish.
func applyFunctionContainer(contSpec *swarm.ContainerSpec, source *entity.DeploymentFunctionSource) {
	contSpec.Dir = ""
	dockerhelper.ContainerCommandApply(contSpec, "")
	contSpec.Healthcheck = nil
	grace := time.Duration(source.Timeout) + functionStopGraceMargin
	contSpec.StopGracePeriod = &grace
}
