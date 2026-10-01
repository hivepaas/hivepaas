package appdeploymentserviceimpl

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
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
	// ContextDir is the build's context when it is not the whole checkout: a
	// function's directory in its repository.
	ContextDir string
}

// contextDir is what the build is given as its context.
func (data *repoDeploymentData) contextDir() string {
	if data.ContextDir != "" {
		return data.ContextDir
	}
	return data.CheckoutDir
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
	err = s.deployStepPrepareBuild(ctx, db, data, data.Deployment.Settings.RepoSource.PushToRegistry)
	if err != nil {
		return hperrors.Wrap(err)
	}

	if data.IsTaskCanceled() {
		return nil
	}

	// 1. Repo checkout
	err = s.repoDeployStepSourceCheckout(ctx, data)
	if err != nil {
		return hperrors.Wrap(err)
	}

	if data.IsTaskCanceled() {
		return nil
	}

	// 2. Build image
	err = s.repoDeployStepImageBuild(ctx, db, data)
	if err != nil {
		return hperrors.Wrap(err)
	}

	if data.IsTaskCanceled() {
		return nil
	}

	// From now until the end of the deployment, we need to lock the app
	// to prevent unexpected behavior in case there are multiple deployments
	// happen at the same time.

	shouldContinue, err := s.lockDockerServiceForDeployment(ctx, db, data.appDeploymentData)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if !shouldContinue {
		data.DeploymentCanceled = true
		return nil
	}

	// 3. Pre-deployment command execution
	err = s.deployStepExecCmd(ctx, data.appDeploymentData, true)
	if err != nil {
		return hperrors.Wrap(err)
	}

	// Pre-deployment jobs, the deploy waiting for those a trigger holds it for
	err = s.deployStepPreDeployJobs(ctx, data.appDeploymentData)
	if err != nil {
		return hperrors.Wrap(err)
	}

	// 4. Apply image to service
	err = s.repoDeployStepServiceApply(ctx, db, data)
	if err != nil {
		return hperrors.Wrap(err)
	}

	// 5. Post-deployment command execution
	err = s.deployStepExecCmd(ctx, data.appDeploymentData, false)
	if err != nil {
		return hperrors.Wrap(err)
	}

	return nil
}

// deployStepPrepareBuild makes the directories a build works in and loads its
// settings. A build on a cluster of several nodes whose image goes to no
// registry is warned about: the other nodes cannot pull it.
func (s *service) deployStepPrepareBuild(
	ctx context.Context,
	db database.IDB,
	data *repoDeploymentData,
	pushToRegistry entity.ObjectID,
) (err error) {
	deployment := data.Deployment

	// Creates temp dir and checkout dir
	data.TempDir, err = fileutil.CreateTempDir(base.BaseTempDirDefault, "*", 0)
	if err != nil {
		return hperrors.Wrap(err)
	}
	data.TempDir, _ = filepath.Abs(data.TempDir)
	data.CheckoutDir = filepath.Join(data.TempDir, "checkout")

	// Load build settings
	err = s.loadImageBuildSettings(ctx, db, data)
	if err != nil {
		return hperrors.Wrap(err)
	}

	// Validate settings
	data.IsMultiNode, err = s.clusterService.IsMultiNode(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if data.IsMultiNode && pushToRegistry.ID == "" {
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
