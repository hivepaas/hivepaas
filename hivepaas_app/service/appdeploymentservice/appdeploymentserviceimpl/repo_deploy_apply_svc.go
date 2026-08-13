package appdeploymentserviceimpl

import (
	"context"
	"errors"
	"time"

	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/dockerhelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

const (
	dockerServiceApplyRetryMax = 2
)

func (s *service) repoDeployStepServiceApply(
	ctx context.Context,
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
		contSpec.Image = data.Deployment.Output.ImageTags[0]
		contSpec.Dir = deployment.Settings.WorkingDir
		dockerhelper.ContainerCommandApply(contSpec, deployment.Settings.Command)

		_, e = s.dockerManager.ServiceUpdate(ctx, data.App.ServiceID, &service.Version, spec,
			func(options *client.ServiceUpdateOptions) {
				options.EncodedRegistryAuth = regAuthHeader
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

	return nil
}
