package appdeploymentserviceimpl

import (
	"context"
	"os"
	"path/filepath"

	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/fileutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagebuildservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/repocheckoutservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

const (
	stepRepoCheckout = "repo-checkout"
	stepImageBuild   = "image-build"
	stepServiceApply = "service-apply"
)

type repoDeploymentData struct {
	*appDeploymentData
	ImageBuildSettings *entity.ImageBuildSettings

	TempDir     string
	CheckoutDir string
}

func (s *service) deployFromRepo(
	ctx context.Context,
	db database.Tx,
	deplData *appDeploymentData,
) error {
	data := &repoDeploymentData{appDeploymentData: deplData}
	data.OnCommand(func(cmd base.TaskCommand, args ...any) {
		s.repoDeployOnCommand(ctx, data, cmd, args...)
	})
	defer s.repoDeployStepCleanup(data) //nolint:errcheck

	// 0. Prepare
	err := s.repoDeployStepPrepare(ctx, db, data)
	if err != nil {
		return apperrors.New(err)
	}

	if data.IsTaskCanceled() {
		return nil
	}

	// 1. Repo checkout
	err = s.repoDeployStepSourceCheckout(ctx, data)
	if err != nil {
		return apperrors.New(err)
	}

	if data.IsTaskCanceled() {
		return nil
	}

	// 2. Build image
	err = s.repoDeployStepImageBuild(ctx, db, data)
	if err != nil {
		return apperrors.New(err)
	}

	if data.IsTaskCanceled() {
		return nil
	}

	// From now until the end of the deployment, we need to lock the app
	// to prevent unexpected behavior in case there are multiple deployments
	// happen at the same time.

	shouldContinue, err := s.lockDockerServiceForDeployment(ctx, db, data.appDeploymentData)
	if err != nil {
		return apperrors.New(err)
	}
	if !shouldContinue {
		data.DeploymentCanceled = true
		return nil
	}

	// 3. Pre-deployment command execution
	err = s.deployStepExecCmd(ctx, data.appDeploymentData, true)
	if err != nil {
		return apperrors.New(err)
	}

	// 4. Apply image to service
	err = s.repoDeployStepServiceApply(ctx, data)
	if err != nil {
		return apperrors.New(err)
	}

	// 5. Post-deployment command execution
	err = s.deployStepExecCmd(ctx, data.appDeploymentData, false)
	if err != nil {
		return apperrors.New(err)
	}

	return nil
}

func (s *service) repoDeployStepSourceCheckout(
	ctx context.Context,
	data *repoDeploymentData,
) (err error) {
	data.Step = stepRepoCheckout
	deployment := data.Deployment
	repoSource := deployment.Settings.RepoSource

	checkoutReq := &repocheckoutservice.RepoCheckoutReq{
		Project:     data.Project,
		App:         data.App,
		RepoSource:  repoSource,
		CredSetting: data.RefObjects.RefSettings[repoSource.Credentials.ID],
		RefObjects:  data.RefObjects,
		LogStore:    data.LogStore,
		TempDir:     data.TempDir,
		CheckoutDir: data.CheckoutDir,
	}
	if deployment.Settings.NoCache || (data.ImageBuildSettings != nil && data.ImageBuildSettings.NoCache) {
		checkoutReq.NoCache = true
	}

	checkoutResp, err := s.repoCheckoutService.Checkout(ctx, checkoutReq)
	if err != nil {
		return apperrors.New(err)
	}

	repoSource.CommitHash = checkoutResp.CommitHash
	data.DeploymentOutput.CommitHash = checkoutResp.CommitHash
	data.DeploymentOutput.CommitMessage = checkoutResp.CommitMessage
	data.DeploymentOutput.CommitTitle = checkoutResp.CommitTitle
	data.DeploymentOutput.CommitAuthor = checkoutResp.CommitAuthor

	return nil
}

func (s *service) repoDeployStepImageBuild(
	ctx context.Context,
	db database.Tx,
	data *repoDeploymentData,
) (err error) {
	data.Step = stepImageBuild
	deployment := data.Deployment

	buildReq := &imagebuildservice.ImageBuildReq{
		Project:            data.Project,
		App:                data.App,
		RepoSource:         deployment.Settings.RepoSource,
		ImageBuildSettings: data.ImageBuildSettings,
		BuildID:            data.Task.ID,
		RefObjects:         data.RefObjects,
		LogStore:           data.LogStore,
		TempDir:            data.TempDir,
		CheckoutDir:        data.CheckoutDir,
	}
	if deployment.Settings.NoCache || (data.ImageBuildSettings != nil && data.ImageBuildSettings.NoCache) {
		buildReq.NoCache = true
	}

	buildResp, err := s.imageBuildService.ImageBuild(ctx, db, buildReq)
	if err != nil {
		return apperrors.New(err)
	}

	data.DeploymentOutput.ImageTags = buildResp.ImageTags

	return nil
}

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
			return apperrors.New(err)
		}
	}

	inspect, err := s.dockerManager.ServiceInspect(ctx, data.App.ServiceID)
	if err != nil {
		return apperrors.New(err)
	}
	service := &inspect.Service
	spec := &service.Spec
	contSpec := spec.TaskTemplate.ContainerSpec
	contSpec.Image = data.DeploymentOutput.ImageTags[0]
	contSpec.Dir = deployment.Settings.WorkingDir
	docker.ContainerCommandApply(contSpec, deployment.Settings.Command)

	_, err = s.dockerManager.ServiceUpdate(ctx, data.App.ServiceID, &service.Version, spec,
		func(options *client.ServiceUpdateOptions) {
			options.EncodedRegistryAuth = regAuthHeader
		})
	if err != nil {
		return apperrors.New(err)
	}

	return nil
}

func (s *service) repoDeployStepPrepare(
	ctx context.Context,
	db database.IDB,
	data *repoDeploymentData,
) (err error) {
	// Creates temp dir and checkout dir
	data.TempDir, err = fileutil.CreateTempDir(base.BaseTempDirDefault, "*", 0)
	if err != nil {
		return apperrors.New(err)
	}
	data.TempDir, _ = filepath.Abs(data.TempDir)
	data.CheckoutDir = filepath.Join(data.TempDir, "checkout")

	// Load build settings
	err = s.loadImageBuildSettings(ctx, db, data)
	if err != nil {
		return apperrors.New(err)
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
