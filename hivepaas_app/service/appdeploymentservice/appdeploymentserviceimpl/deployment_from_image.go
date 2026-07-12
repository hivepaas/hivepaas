package appdeploymentserviceimpl

import (
	"context"
	"errors"

	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/batchrecvchan"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/services/docker"
)

const (
	stepImagePull = "image-pull"
)

type imageDeploymentData struct {
	*appDeploymentData
	RegAuthHeader string
	Step          string
}

func (s *service) deployFromImage(
	ctx context.Context,
	db database.Tx,
	deplData *appDeploymentData,
) error {
	data := &imageDeploymentData{appDeploymentData: deplData}

	// 1. Pull image from the registry
	err := s.imageDeployStepImagePull(ctx, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	if data.IsTaskCanceled() {
		return nil
	}

	// From now until the end of the deployment, we need to lock the app
	// to prevent unexpected behavior in case there are multiple deployments
	// happen at the same time.

	shouldContinue, err := s.lockDockerServiceForDeployment(ctx, db, data.appDeploymentData)
	if err != nil {
		return apperrors.Wrap(err)
	}
	if !shouldContinue {
		data.DeploymentCanceled = true
		return nil
	}

	// 2. Pre-deployment command execution
	err = s.deployStepExecCmd(ctx, data.appDeploymentData, true)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// 3. Apply image to service
	err = s.imageDeployStepServiceApply(ctx, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// 4. Post-deployment command execution
	err = s.deployStepExecCmd(ctx, data.appDeploymentData, false)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}

func (s *service) imageDeployStepImagePull(
	ctx context.Context,
	data *imageDeploymentData,
) (err error) {
	data.Step = stepImagePull
	imageSource := data.Deployment.Settings.ImageSource

	s.addStepStartLog(ctx, data.appDeploymentData, "Start pulling image...")
	defer s.addStepEndLog(ctx, data.appDeploymentData, timeutil.NowUTC(), err)

	if imageSource.RegistryAuth.ID != "" {
		regAuth := data.RefObjects.RefSettings[imageSource.RegistryAuth.ID]
		data.RegAuthHeader, err = regAuth.MustAsRegistryAuth().GenerateAuthHeader()
		if err != nil {
			return apperrors.Wrap(err)
		}
	}

	logsReader, err := s.dockerManager.ImagePull(ctx, imageSource.Image, func(options *client.ImagePullOptions) {
		options.RegistryAuth = data.RegAuthHeader
	})
	if err != nil {
		return apperrors.Wrap(err)
	}

	logsChan, _ := docker.StartScanningJSONMsg(ctx, logsReader, batchrecvchan.Options{})
	for msgs := range logsChan {
		for _, msg := range msgs {
			frameCreator := tasklog.NewDebugFrame
			if msg.Error != nil {
				err = errors.Join(err, msg.Error)
				frameCreator = tasklog.NewErrFrame
			}
			if msg.String() != "" {
				_ = data.LogStore.Add(ctx, frameCreator(msg.String(), tasklog.TsNow))
			}
		}
	}
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}

func (s *service) imageDeployStepServiceApply(
	ctx context.Context,
	data *imageDeploymentData,
) (err error) {
	data.Step = stepServiceApply
	deployment := data.Deployment
	imageSource := deployment.Settings.ImageSource

	s.addStepStartLog(ctx, data.appDeploymentData, "Applying changes to service...")
	defer s.addStepEndLog(ctx, data.appDeploymentData, timeutil.NowUTC(), err)

	inspect, err := s.dockerManager.ServiceInspect(ctx, data.App.ServiceID)
	if err != nil {
		return apperrors.Wrap(err)
	}

	service := &inspect.Service
	spec := &service.Spec
	contSpec := spec.TaskTemplate.ContainerSpec
	contSpec.Image = imageSource.Image
	contSpec.Dir = deployment.Settings.WorkingDir
	docker.ContainerCommandApply(contSpec, deployment.Settings.Command)

	_, err = s.dockerManager.ServiceUpdate(ctx, data.App.ServiceID, &service.Version, spec,
		func(options *client.ServiceUpdateOptions) {
			options.EncodedRegistryAuth = data.RegAuthHeader
		})
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Save the used image in the output
	data.DeploymentOutput.ImageTags = append(data.DeploymentOutput.ImageTags, imageSource.Image)

	return nil
}
