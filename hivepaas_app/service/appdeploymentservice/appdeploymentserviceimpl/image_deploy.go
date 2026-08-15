package appdeploymentserviceimpl

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
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
) (err error) {
	data := &imageDeploymentData{appDeploymentData: deplData}
	defer func() {
		if data.IsTaskCanceled() || errors.Is(err, context.Canceled) {
			err = nil
		}
	}()

	// 1. Pull image from the registry
	err = s.imageDeployStepImagePull(ctx, data)
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
	err = s.imageDeployStepServiceApply(ctx, db, data)
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
