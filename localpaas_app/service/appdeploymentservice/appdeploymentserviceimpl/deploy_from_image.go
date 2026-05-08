package appdeploymentserviceimpl

import (
	"context"
	"errors"

	"github.com/moby/moby/client"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/pkg/applog"
	"github.com/localpaas/localpaas/localpaas_app/pkg/batchrecvchan"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/service/appdeploymentservice"
	"github.com/localpaas/localpaas/services/docker"
)

const (
	stepImagePull = "image-pull"
)

type imageDeployTaskData struct {
	*appdeploymentservice.DeploymentData
	RegAuthHeader string
	Step          string
}

func (s *service) deployFromImage(
	ctx context.Context,
	taskData *appdeploymentservice.DeploymentData,
) error {
	data := &imageDeployTaskData{DeploymentData: taskData}

	// 1. Pull image from the registry
	err := s.imageDeployStepImagePull(ctx, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	if data.IsCanceled() {
		return nil
	}

	// 2. Pre-deployment command execution
	err = s.deployStepExecCmd(ctx, data.DeploymentData, true)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// 3. Apply image to service
	err = s.imageDeployStepServiceApply(ctx, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// 4. Post-deployment command execution
	err = s.deployStepExecCmd(ctx, data.DeploymentData, false)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}

func (s *service) imageDeployStepImagePull(
	ctx context.Context,
	data *imageDeployTaskData,
) (err error) {
	data.Step = stepImagePull
	imageSource := data.Deployment.Settings.ImageSource

	s.addStepStartLog(ctx, data.DeploymentData, "Start pulling image...")
	defer s.addStepEndLog(ctx, data.DeploymentData, timeutil.NowUTC(), err)

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
			frameCreator := applog.NewOutFrame
			if msg.Error != nil {
				err = errors.Join(err, msg.Error)
				frameCreator = applog.NewErrFrame
			}
			if msg.String() != "" {
				_ = data.LogStore.Add(ctx, frameCreator(msg.String(), applog.TsNow))
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
	data *imageDeployTaskData,
) (err error) {
	data.Step = stepServiceApply
	deployment := data.Deployment
	imageSource := deployment.Settings.ImageSource

	s.addStepStartLog(ctx, data.DeploymentData, "Applying changes to service...")
	defer s.addStepEndLog(ctx, data.DeploymentData, timeutil.NowUTC(), err)

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

	return nil
}
