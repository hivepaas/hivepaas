package appdeploymentserviceimpl

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/fileutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
)

const (
	stepRepoCheckout = "repo-checkout"
	stepImageBuild   = "image-build"
	stepServiceApply = "service-apply"
)

type repoDeploymentData struct {
	*appDeploymentData
	ImageBuildSettings *entity.ImageBuildSettings
	IsMultiNode        bool

	TempDir     string
	CheckoutDir string
}

func (s *service) deployFromRepo(
	ctx context.Context,
	db database.Tx,
	deplData *appDeploymentData,
) (err error) {
	data := &repoDeploymentData{appDeploymentData: deplData}
	data.OnCommand(func(cmd base.TaskCommand, args ...any) {
		s.repoDeployOnCommand(ctx, data, cmd, args...)
	})
	defer s.repoDeployStepCleanup(data) //nolint:errcheck
	defer func() {
		if data.IsTaskCanceled() || errors.Is(err, context.Canceled) {
			err = nil
		}
	}()

	// 0. Prepare
	err = s.repoDeployStepPrepare(ctx, db, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	if data.IsTaskCanceled() {
		return nil
	}

	// 1. Repo checkout
	err = s.repoDeployStepSourceCheckout(ctx, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	if data.IsTaskCanceled() {
		return nil
	}

	// 2. Build image
	err = s.repoDeployStepImageBuild(ctx, db, data)
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

	// 3. Pre-deployment command execution
	err = s.deployStepExecCmd(ctx, data.appDeploymentData, true)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// 4. Apply image to service
	err = s.repoDeployStepServiceApply(ctx, db, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// 5. Post-deployment command execution
	err = s.deployStepExecCmd(ctx, data.appDeploymentData, false)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}

func (s *service) repoDeployStepPrepare(
	ctx context.Context,
	db database.IDB,
	data *repoDeploymentData,
) (err error) {
	deployment := data.Deployment

	// Creates temp dir and checkout dir
	data.TempDir, err = fileutil.CreateTempDir(base.BaseTempDirDefault, "*", 0)
	if err != nil {
		return apperrors.Wrap(err)
	}
	data.TempDir, _ = filepath.Abs(data.TempDir)
	data.CheckoutDir = filepath.Join(data.TempDir, "checkout")

	// Load build settings
	err = s.loadImageBuildSettings(ctx, db, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Validate settings
	data.IsMultiNode, err = s.clusterService.IsMultiNode(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}
	if data.IsMultiNode && deployment.Settings.RepoSource.PushToRegistry.ID == "" {
		warn := "[WARN] The cluster is multi-node, but no target registry is configured to push the built image. " +
			"The image will not be accessible from other nodes in the cluster."
		deployment.Output.Errors = append(deployment.Output.Errors, warn)
		_ = data.LogStore.Add(ctx, tasklog.NewWarnFrame(warn, tasklog.TsNow))
	}

	return nil
}

//nolint:unparam
func (s *service) repoDeployStepCleanup(
	data *repoDeploymentData,
) (err error) {
	if data.TempDir != "" {
		_ = os.RemoveAll(data.TempDir)
	}
	return nil
}

func (s *service) repoDeployOnCommand(
	_ context.Context,
	data *repoDeploymentData,
	cmd base.TaskCommand,
	_ ...any,
) {
	if cmd == base.TaskCommandCancel && data.Step == stepImageBuild { //nolint
		// TODO: cancel image build
	}
}
