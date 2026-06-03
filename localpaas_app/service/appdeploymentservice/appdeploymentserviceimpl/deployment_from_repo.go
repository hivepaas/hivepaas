package appdeploymentserviceimpl

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/moby/go-archive"
	"github.com/moby/moby/api/types/build"
	"github.com/moby/moby/client"
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/batchrecvchan"
	"github.com/localpaas/localpaas/localpaas_app/pkg/fileutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/githelper"
	"github.com/localpaas/localpaas/localpaas_app/pkg/tasklog"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/services/docker"
)

const (
	stepCodeCheckout = "code-checkout"
	stepImageBuild   = "image-build"
	stepImagePush    = "image-push"
	stepServiceApply = "service-apply"
)

type repoDeploymentData struct {
	*appDeploymentData
	CredSetting        *entity.Setting
	RegAuthHeader      string
	ImageBuildSettings *entity.ImageBuildSettings

	RepoCache        *entity.File
	RepoCacheLoaded  bool
	CheckoutDuration time.Duration

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

	// 3. Push image to a registry if configured
	err = s.repoDeployStepImagePush(ctx, data)
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

	// 4. Pre-deployment command execution
	err = s.deployStepExecCmd(ctx, data.appDeploymentData, true)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// 5. Apply image to service
	err = s.repoDeployStepServiceApply(ctx, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// 6. Post-deployment command execution
	err = s.deployStepExecCmd(ctx, data.appDeploymentData, false)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}

func (s *service) repoDeployStepSourceCheckout(
	ctx context.Context,
	data *repoDeploymentData,
) (err error) {
	data.Step = stepCodeCheckout
	deployment := data.Deployment
	repoSource := deployment.Settings.RepoSource

	// NOTE: currently supports repo of git type only
	if repoSource.RepoType != base.RepoTypeGit {
		_ = data.LogStore.Add(ctx, tasklog.NewErrFrame("Failed to checkout source: "+
			"unsupported repository type: "+string(repoSource.RepoType), tasklog.TsNow))
		return apperrors.NewUnsupported(apperrors.Fmt("Repository type '%v'", repoSource.RepoType))
	}

	s.addStepStartLog(ctx, data.appDeploymentData, "Start cloning Git repository...")
	defer s.addStepEndLog(ctx, data.appDeploymentData, timeutil.NowUTC(), err)

	err = s.repoCheckoutLoadCache(ctx, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	authMethod, err := s.calcGitAuthMethod(ctx, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	recurseSubmodules := gofn.If(repoSource.RepoOptions.GitSubmodulesEnabled, git.DefaultSubmoduleRecursionDepth,
		git.NoRecurseSubmodules)
	lfsEnabled := repoSource.RepoOptions.GitLFSEnabled

	checkoutMaxDepth := uint(0)
	if data.ImageBuildSettings != nil {
		checkoutMaxDepth = data.ImageBuildSettings.Sources.CheckoutMaxDepth
	}

	checkoutOptions := &githelper.CheckoutOptions{
		URL:           repoSource.RepoURL,
		ReferenceName: plumbing.ReferenceName(repoSource.RepoRef),
		Auth:          authMethod,
		SingleBranch:  true,
		CommitHash:    repoSource.CommitHash,

		Depth:    1,
		MaxDepth: checkoutMaxDepth,

		RecurseSubmodules: recurseSubmodules,
		ShallowSubmodules: true,

		LFSEnabled: lfsEnabled,

		TempDir:     data.TempDir,
		CheckoutDir: data.CheckoutDir,
		CacheLoaded: data.RepoCacheLoaded,
		LogStore:    data.LogStore,
	}

	var commit *object.Commit
	checkoutStart := time.Now()
	for {
		_, commit, err = githelper.CheckoutWithGitCli(ctx, checkoutOptions)
		if err == nil {
			break
		}
		if checkoutOptions.CacheLoaded {
			if err := s.resetRepoCheckoutDir(data); err != nil {
				return apperrors.Wrap(err)
			}
			_ = data.LogStore.Add(ctx, tasklog.NewWarnFrame("Failed to checkout repository with error: "+
				err.Error()+". Try to do a fresh clone (not using cache)...", tasklog.TsNow))
			checkoutOptions.CacheLoaded = false
			data.RepoCacheLoaded = false
			continue
		}
		_ = data.LogStore.Add(ctx, tasklog.NewErrFrame("Failed to checkout repository with error: "+
			err.Error(), tasklog.TsNow))
		return apperrors.Wrap(err)
	}

	data.CheckoutDuration = time.Since(checkoutStart)
	data.DeploymentOutput.CommitHash = commit.Hash.String()
	data.DeploymentOutput.CommitMessage = commit.Message

	// Cache the latest repo source if satisfied our condition
	ee := s.repoCheckoutSaveCache(ctx, data)
	if ee != nil { // Just log
		_ = data.LogStore.Add(ctx, tasklog.NewErrFrame("Failed to cache repository source: "+
			ee.Error(), tasklog.TsNow))
	}

	// Remove .git dir within the source dir before building image
	ee = os.RemoveAll(filepath.Join(data.CheckoutDir, ".git"))
	if ee != nil { // Just log
		_ = data.LogStore.Add(ctx, tasklog.NewErrFrame("Failed to remove .git directory from source: "+
			ee.Error(), tasklog.TsNow))
	}

	return nil
}

func (s *service) repoDeployStepImageBuild(
	ctx context.Context,
	db database.Tx,
	data *repoDeploymentData,
) (err error) {
	data.Step = stepImageBuild
	deployment := data.Deployment
	repoSource := deployment.Settings.RepoSource
	buildSetting := data.ImageBuildSettings

	s.addStepStartLog(ctx, data.appDeploymentData, "Start building image...")
	defer s.addStepEndLog(ctx, data.appDeploymentData, timeutil.NowUTC(), err)

	// TODO: check dockerfile existence
	dockerfile := gofn.Coalesce(repoSource.DockerfilePath, "Dockerfile")

	imageTags, err := s.calcBuildImageTags(repoSource.ImageTags, data)
	if err != nil {
		return apperrors.Wrap(err)
	}
	data.DeploymentOutput.ImageTags = imageTags

	envVars, err := s.calcBuildEnvVars(ctx, db, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	authConfigs, err := s.calcBuildRegistryAuths(ctx, db, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Create tar archive for the source code
	tar, err := archive.TarWithOptions(data.CheckoutDir, &archive.TarOptions{})
	if err != nil {
		return apperrors.Wrap(err)
	}
	defer tar.Close()

	// Build the image
	resp, err := s.dockerManager.ImageBuild(ctx, tar, func(opts *client.ImageBuildOptions) {
		opts.Version = build.BuilderV1
		opts.BuildID = data.Task.ID
		opts.Dockerfile = dockerfile
		opts.Tags = imageTags
		opts.BuildArgs = envVars
		opts.AuthConfigs = authConfigs

		if buildSetting != nil {
			opts.NoCache = buildSetting.NoCache
			opts.SuppressOutput = buildSetting.NoVerbose
			res := buildSetting.Resources
			if res.CPUs > 0 {
				opts.CPUPeriod, opts.CPUQuota = res.CPUsAsPeriodAndQuota()
			}
			if res.Mem > 0 {
				opts.Memory = res.Mem.Bytes()
			}
			if res.MemSwap > 0 {
				opts.MemorySwap = res.MemSwap.Bytes()
			}
			if res.ShmSize > 0 {
				opts.ShmSize = res.ShmSize.Bytes()
			}
		}
	})
	if err != nil {
		return apperrors.Wrap(err)
	}

	logsChan, _ := docker.StartScanningJSONMsg(ctx, resp.Body, batchrecvchan.Options{})
	for msgs := range logsChan {
		for _, msg := range msgs {
			frameCreator := tasklog.NewOutFrame
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

func (s *service) repoDeployStepImagePush(
	ctx context.Context,
	data *repoDeploymentData,
) (err error) {
	deployment := data.Deployment
	repoSource := deployment.Settings.RepoSource
	if repoSource.PushToRegistry.ID == "" {
		return nil
	}
	data.Step = stepImagePush

	s.addStepStartLog(ctx, data.appDeploymentData, "Start pushing image to registry...")
	defer s.addStepEndLog(ctx, data.appDeploymentData, timeutil.NowUTC(), err)

	regAuth := data.RefObjects.RefSettings[repoSource.PushToRegistry.ID]
	data.RegAuthHeader, err = regAuth.MustAsRegistryAuth().GenerateAuthHeader()
	if err != nil {
		return apperrors.Wrap(err)
	}

	for _, tag := range data.DeploymentOutput.ImageTags {
		if !strings.Contains(tag, "/") { // only push tag containing `/` in it
			continue
		}
		logsReader, err := s.dockerManager.ImagePush(ctx, tag, func(options *client.ImagePushOptions) {
			options.RegistryAuth = data.RegAuthHeader
		})
		if err != nil {
			return apperrors.Wrap(err)
		}

		logsChan, _ := docker.StartScanningJSONMsg(ctx, logsReader, batchrecvchan.Options{})
		for msgs := range logsChan {
			for _, msg := range msgs {
				frameCreator := tasklog.NewOutFrame
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
	}

	return nil
}

func (s *service) repoDeployStepServiceApply(
	ctx context.Context,
	data *repoDeploymentData,
) (err error) {
	data.Step = stepServiceApply
	deployment := data.Deployment

	s.addStepStartLog(ctx, data.appDeploymentData, "Applying changes to service...")
	defer s.addStepEndLog(ctx, data.appDeploymentData, timeutil.NowUTC(), err)

	inspect, err := s.dockerManager.ServiceInspect(ctx, data.App.ServiceID)
	if err != nil {
		return apperrors.Wrap(err)
	}
	service := &inspect.Service
	spec := &service.Spec
	contSpec := spec.TaskTemplate.ContainerSpec
	contSpec.Image = data.DeploymentOutput.ImageTags[0]
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

func (s *service) repoDeployStepPrepare(
	ctx context.Context,
	db database.IDB,
	data *repoDeploymentData,
) (err error) {
	deployment := data.Deployment
	repoSource := deployment.Settings.RepoSource

	// Loads repo credentials (github app, git token, ssh key) if configured
	if repoSource.Credentials.ID != "" {
		data.CredSetting = data.RefObjects.RefSettings[repoSource.Credentials.ID]
	}

	// Creates temp dir and checkout dir
	data.TempDir, err = fileutil.CreateTempDir(base.BaseTempDirDefault, "*", 0)
	if err != nil {
		return apperrors.Wrap(err)
	}
	data.TempDir, _ = filepath.Abs(data.TempDir)
	data.CheckoutDir = filepath.Join(data.TempDir, "checkout")

	err = os.MkdirAll(data.CheckoutDir, base.DirModeDefault)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Load build settings
	err = s.loadImageBuildSettings(ctx, db, data)
	if err != nil {
		return apperrors.Wrap(err)
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
	ctx context.Context,
	data *repoDeploymentData,
	cmd base.TaskCommand,
	_ ...any,
) {
	if cmd == base.TaskCommandCancel && data.Step == stepImageBuild {
		_, err := s.dockerManager.ImageBuildCancel(ctx, data.Task.ID)
		if err != nil {
			_ = data.LogStore.Add(ctx, tasklog.NewErrFrame("failed to cancel image build: "+
				err.Error(), tasklog.TsNow))
		}
	}
}
