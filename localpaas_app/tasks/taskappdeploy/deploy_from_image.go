package taskappdeploy

import (
	"context"
	"errors"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/swarm"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/pkg/applog"
	"github.com/localpaas/localpaas/localpaas_app/pkg/batchrecvchan"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/services/docker"
)

const (
	stepImagePull = "image-pull"
)

type imageDeployTaskData struct {
	*taskData
	RegAuthHeader string
	Step          string
}

func (e *Executor) deployFromImage(
	ctx context.Context,
	taskData *taskData,
) error {
	data := &imageDeployTaskData{taskData: taskData}

	// 1. Pull image from the registry
	err := e.imageDeployStepImagePull(ctx, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	if data.IsCanceled() {
		return nil
	}

	// 2. Pre-deployment command execution
	err = e.deployStepExecCmd(ctx, data.taskData, true)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// 3. Apply image to service
	err = e.imageDeployStepServiceApply(ctx, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// 4. Post-deployment command execution
	err = e.deployStepExecCmd(ctx, data.taskData, false)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}

func (e *Executor) imageDeployStepImagePull(
	ctx context.Context,
	data *imageDeployTaskData,
) (err error) {
	data.Step = stepImagePull
	imageSource := data.Deployment.Settings.ImageSource

	e.addStepStartLog(ctx, data.taskData, "Start pulling image...")
	defer e.addStepEndLog(ctx, data.taskData, timeutil.NowUTC(), err)

	if imageSource.RegistryAuth.ID != "" {
		regAuth := data.RefObjects.RefSettings[imageSource.RegistryAuth.ID]
		data.RegAuthHeader, err = regAuth.MustAsRegistryAuth().GenerateAuthHeader()
		if err != nil {
			return apperrors.Wrap(err)
		}
	}

	logsReader, err := e.dockerManager.ImagePull(ctx, imageSource.Image, func(options *image.PullOptions) {
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

func (e *Executor) imageDeployStepServiceApply(
	ctx context.Context,
	data *imageDeployTaskData,
) (err error) {
	data.Step = stepServiceApply
	deployment := data.Deployment
	imageSource := deployment.Settings.ImageSource

	e.addStepStartLog(ctx, data.taskData, "Applying changes to service...")
	defer e.addStepEndLog(ctx, data.taskData, timeutil.NowUTC(), err)

	service, err := e.dockerManager.ServiceInspect(ctx, data.App.ServiceID)
	if err != nil {
		return apperrors.Wrap(err)
	}

	spec := &service.Spec
	contSpec := spec.TaskTemplate.ContainerSpec
	contSpec.Image = imageSource.Image
	contSpec.Dir = deployment.Settings.WorkingDir
	docker.ContainerCommandApply(contSpec, deployment.Settings.Command)

	_, err = e.dockerManager.ServiceUpdate(ctx, data.App.ServiceID, &service.Version, spec,
		func(options *swarm.ServiceUpdateOptions) {
			options.EncodedRegistryAuth = data.RegAuthHeader
		})
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}
