package appdeploymentserviceimpl

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/functionservice/functionbuild"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagebuildservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/repocheckoutservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/systemappservice"
)

const (
	stepFunctionSource = "function-source"
)

// deployFromFunction deploys a function: its code, inline or from its
// repository, is built on its runtime's image with the Dockerfile HivePaaS
// writes for it, then runs as the app's service.
func (s *service) deployFromFunction(
	ctx context.Context,
	db database.Tx,
	deplData *appDeploymentData,
) (err error) {
	data := &repoDeploymentData{appDeploymentData: deplData}
	defer s.repoDeployStepCleanup(data) //nolint:errcheck
	defer func() {
		if data.IsTaskCanceled() || errors.Is(err, context.Canceled) {
			err = nil
		}
	}()
	source := data.Deployment.Settings.FunctionSource
	if source == nil {
		return hperrors.Wrap(hperrors.ErrUnconfigured).WithParam("Name", "Function source")
	}

	// 0. Prepare
	if err = s.deployStepPrepareBuild(ctx, db, data, source.PushToRegistry); err != nil {
		return hperrors.Wrap(err)
	}
	if data.IsTaskCanceled() {
		return nil
	}

	// 1. The function's source
	if err = s.functionDeployStepSource(ctx, data); err != nil {
		return hperrors.Wrap(err)
	}
	if data.IsTaskCanceled() {
		return nil
	}

	// 2. Build image
	if err = s.functionDeployStepImageBuild(ctx, db, data); err != nil {
		return hperrors.Wrap(err)
	}
	if data.IsTaskCanceled() {
		return nil
	}

	// From now until the end of the deployment, the app is locked against
	// another deployment, as a repository's is.
	shouldContinue, err := s.lockDockerServiceForDeployment(ctx, db, data.appDeploymentData)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if !shouldContinue {
		data.DeploymentCanceled = true
		return nil
	}

	// 3. Pre-deployment command execution, and the jobs waited for
	if err = s.deployStepExecCmd(ctx, data.appDeploymentData, true); err != nil {
		return hperrors.Wrap(err)
	}
	if err = s.deployStepPreDeployJobs(ctx, data.appDeploymentData); err != nil {
		return hperrors.Wrap(err)
	}

	// 4. Apply image to service
	if err = s.functionDeployStepServiceApply(ctx, db, data); err != nil {
		return hperrors.Wrap(err)
	}

	// 5. Post-deployment command execution
	if err = s.deployStepExecCmd(ctx, data.appDeploymentData, false); err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

// functionDeployStepSource puts the function's code in the checkout directory:
// its inline files, or its repository checked out. The build's context is the
// function's own directory.
func (s *service) functionDeployStepSource(
	ctx context.Context,
	data *repoDeploymentData,
) error {
	data.Step = stepFunctionSource
	code := data.Deployment.Settings.FunctionSource.Code

	switch {
	case code.Inline != nil:
		_ = data.LogStore.Add(ctx, tasklog.NewOutFrame(
			fmt.Sprintf("Writing the function's %d file(s)...", len(code.Inline.Files)), tasklog.TsNow))
		if err := os.MkdirAll(data.CheckoutDir, 0o700); err != nil { //nolint:mnd
			return hperrors.Wrap(err)
		}
		if err := functionbuild.WriteInlineCode(data.CheckoutDir, code.Inline); err != nil {
			return hperrors.Wrap(err)
		}
		data.ContextDir = data.CheckoutDir

	case code.Repo != nil:
		repoSource := code.Repo.RepoSource()
		checkoutReq := &repocheckoutservice.RepoCheckoutReq{
			App:         data.App,
			RepoSource:  repoSource,
			CredSetting: data.RefObjects.RefSettings[repoSource.Credentials.ID],
			RefObjects:  data.RefObjects,
			LogStore:    data.LogStore,
			TempDir:     data.TempDir,
			CheckoutDir: data.CheckoutDir,
			NoCache: data.DeployArgs.NoCache ||
				(data.ImageBuildSettings != nil && data.ImageBuildSettings.NoCache),
		}
		checkoutResp, err := s.repoCheckoutService.Checkout(ctx, checkoutReq)
		if err != nil {
			return hperrors.Wrap(err)
		}
		code.Repo.CommitHash = checkoutResp.CommitHash
		output := data.Deployment.Output
		output.CommitHash = checkoutResp.CommitHash
		output.CommitMessage = checkoutResp.CommitMessage
		output.CommitTitle = checkoutResp.CommitTitle
		output.CommitAuthor = checkoutResp.CommitAuthor

		data.ContextDir = filepath.Join(data.CheckoutDir, filepath.FromSlash(code.Dir))
		if info, err := os.Stat(data.ContextDir); err != nil || !info.IsDir() {
			_ = data.LogStore.Add(ctx, tasklog.NewErrFrame(fmt.Sprintf(
				"The repository has no directory '%s' at commit %s.", code.Dir, checkoutResp.CommitHash),
				tasklog.TsNow))
			return hperrors.NewNotFound("Function directory")
		}

	default:
		return hperrors.Wrap(hperrors.ErrUnconfigured).WithParam("Name", "Function code")
	}
	return nil
}

// functionDeployStepImageBuild builds the function from the Dockerfile written
// for it. The build's inputs are resolved first: the Dockerfile mounts each of
// the build's secrets into the install step, so it needs their names.
func (s *service) functionDeployStepImageBuild(
	ctx context.Context,
	db database.Tx,
	data *repoDeploymentData,
) error {
	source := data.Deployment.Settings.FunctionSource
	inputs, err := s.imageBuildService.ResolveBuildInputs(ctx, db, &imagebuildservice.ImageBuildReq{
		App:            data.App,
		PushToRegistry: source.PushToRegistry,
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	data.LogStore.UpdateRedactorAddSecrets(inputs.Secrets)

	dockerfile, err := functionbuild.Dockerfile(&functionbuild.DockerfileReq{
		Source:       source,
		Images:       systemappservice.CurrentRelease().FunctionRuntimes,
		SourceDir:    data.contextDir(),
		BuildArgs:    gofn.MapKeys(inputs.EnvVars),
		BuildSecrets: gofn.MapKeys(inputs.SecretEnvVars),
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	for _, note := range dockerfile.Notes {
		_ = data.LogStore.Add(ctx, tasklog.NewWarnFrame(note, tasklog.TsNow))
	}
	data.Deployment.Output.RuntimeImages = dockerfile.Images

	// Inline code has no commit: its image is tagged after its content.
	commitHash := ""
	if source.Code.Repo != nil {
		commitHash = source.Code.Repo.CommitHash
	} else if source.Code.Inline != nil {
		commitHash = functionbuild.ContentHash(source.Code.Inline, dockerfile.Content)
	}

	return s.deployStepImageBuild(ctx, db, data, &imageBuildTarget{
		CommitHash: commitHash,
		Dockerfile: entity.DeploymentDockerfile{
			Source:  base.DockerfileSourceManual,
			Path:    functionbuild.DockerfilePath,
			Content: dockerfile.Content,
		},
		PushToRegistry: source.PushToRegistry,
		Inputs:         inputs,
	})
}
