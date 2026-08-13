package appdeploymentserviceimpl

import (
	"context"
	"errors"

	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/batchrecvchan"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/services/docker"
)

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
