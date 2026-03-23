package taskappdeploy

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types/build"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/swarm"
	"github.com/gitsight/go-vcsurl"
	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/moby/go-archive"
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/applog"
	"github.com/localpaas/localpaas/localpaas_app/pkg/batchrecvchan"
	"github.com/localpaas/localpaas/localpaas_app/pkg/fileutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/services/docker"
)

const (
	stepCodeCheckout = "code-checkout"
	stepImageBuild   = "image-build"
	stepImagePush    = "image-push"
	stepServiceApply = "service-apply"
)

type repoDeployTaskData struct {
	*taskData
	CredSetting   *entity.Setting
	CheckoutPath  string
	RepoURLInfo   *vcsurl.VCS
	RegAuthHeader string
}

func (e *Executor) deployFromRepo(
	ctx context.Context,
	db database.Tx,
	taskData *taskData,
) error {
	data := &repoDeployTaskData{taskData: taskData}
	data.OnCommand(func(cmd base.TaskCommand, args ...any) { e.onRepoDeployCommand(ctx, data, cmd, args...) })

	// 0. Prepare
	err := e.repoDeployStepPrepare(ctx, data)
	if err != nil {
		return apperrors.Wrap(err)
	}
	defer e.repoDeployStepCleanup(ctx, data) //nolint:errcheck

	if data.IsCanceled() {
		return nil
	}

	// 1. Repo checkout
	err = e.repoDeployStepSourceCheckout(ctx, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	if data.IsCanceled() {
		return nil
	}

	// 2. Build image
	err = e.repoDeployStepImageBuild(ctx, db, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	if data.IsCanceled() {
		return nil
	}

	// 3. Push image to a registry if configured
	err = e.repoDeployStepImagePush(ctx, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	if data.IsCanceled() {
		return nil
	}

	// 4. Pre-deployment command execution
	err = e.deployStepExecCmd(ctx, data.taskData, true)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// 5. Apply image to service
	err = e.repoDeployStepServiceApply(ctx, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// 6. Post-deployment command execution
	err = e.deployStepExecCmd(ctx, data.taskData, false)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}

func (e *Executor) repoDeployStepSourceCheckout(
	ctx context.Context,
	data *repoDeployTaskData,
) (err error) {
	data.Step = stepCodeCheckout
	deployment := data.Deployment
	repoSource := deployment.Settings.RepoSource

	// NOTE: currently supports repo of git type only
	if repoSource.RepoType != base.RepoTypeGit {
		return apperrors.New(apperrors.ErrUnsupported).
			WithExtraDetail("Repo type %s is unsupported", repoSource.RepoType)
	}

	e.addStepStartLog(ctx, data.taskData, "Start cloning Git repository...")
	defer e.addStepEndLog(ctx, data.taskData, timeutil.NowUTC(), err)

	authMethod, err := e.calcGitAuthMethod(ctx, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	repo, err := git.PlainCloneContext(ctx, data.CheckoutPath, &git.CloneOptions{
		URL:               repoSource.RepoURL,
		ReferenceName:     plumbing.ReferenceName(repoSource.RepoRef),
		Auth:              authMethod,
		Depth:             1,
		SingleBranch:      true,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		ShallowSubmodules: true,
	})
	if err != nil {
		return apperrors.Wrap(err)
	}

	commitIter, err := repo.CommitObjects()
	if err != nil {
		return apperrors.Wrap(err)
	}
	commit, err := commitIter.Next()
	if err != nil {
		return apperrors.Wrap(err)
	}
	data.DeploymentOutput.CommitHash = commit.Hash.String()
	data.DeploymentOutput.CommitMessage = commit.Message

	// Remove .git dir within the source dir
	ee := os.RemoveAll(filepath.Join(data.CheckoutPath, ".git"))
	if ee != nil { // Just log
		_ = data.LogStore.Add(ctx, applog.NewErrFrame("failed to remove .git folder",
			applog.TsNow))
	}

	return nil
}

//nolint:gocognit
func (e *Executor) repoDeployStepImageBuild(
	ctx context.Context,
	db database.Tx,
	data *repoDeployTaskData,
) (err error) {
	data.Step = stepImageBuild
	deployment := data.Deployment
	repoSource := deployment.Settings.RepoSource

	e.addStepStartLog(ctx, data.taskData, "Start building image...")
	defer e.addStepEndLog(ctx, data.taskData, timeutil.NowUTC(), err)

	// TODO: check dockerfile existence
	dockerfile := gofn.Coalesce(repoSource.DockerfilePath, "Dockerfile")

	buildSetting, err := e.getBuildSetting(ctx, db, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	imageTags, err := e.calcBuildImageTags(repoSource.ImageTags, data)
	if err != nil {
		return apperrors.Wrap(err)
	}
	data.DeploymentOutput.ImageTags = imageTags

	envVars, err := e.calcBuildEnvVars(ctx, db, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	authConfigs, err := e.calcBuildRegistryAuths(ctx, db, data)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Create tar archive for the source code
	tar, err := archive.TarWithOptions(data.CheckoutPath, &archive.TarOptions{})
	if err != nil {
		return apperrors.Wrap(err)
	}
	defer tar.Close()

	// Build the image
	resp, err := e.dockerManager.ImageBuild(ctx, tar, func(opts *build.ImageBuildOptions) {
		opts.Version = build.BuilderV1
		opts.BuildID = data.Task.ID
		opts.Dockerfile = dockerfile
		opts.Tags = imageTags
		opts.BuildArgs = envVars
		opts.AuthConfigs = authConfigs

		if buildSetting != nil { //nolint:nestif
			opts.NoCache = buildSetting.NoCache
			opts.SuppressOutput = buildSetting.NoVerbose
			if buildSetting.Resources != nil {
				res := buildSetting.Resources
				if res.CPUs > 0 {
					opts.CPUPeriod, opts.CPUQuota = res.CPUsAsPeriodAndQuota()
				}
				if res.MemMB > 0 {
					opts.Memory = res.MemMB * docker.UnitMemMB
				}
				if res.MemSwapMB > 0 {
					opts.MemorySwap = res.MemSwapMB * docker.UnitMemMB
				}
				if res.ShmSizeMB > 0 {
					opts.ShmSize = res.ShmSizeMB * docker.UnitMemMB
				}
			}
		}
	})
	if err != nil {
		return apperrors.Wrap(err)
	}

	logsChan, _ := docker.StartScanningJSONMsg(ctx, resp.Body, batchrecvchan.Options{})
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

func (e *Executor) repoDeployStepImagePush(
	ctx context.Context,
	data *repoDeployTaskData,
) (err error) {
	deployment := data.Deployment
	repoSource := deployment.Settings.RepoSource
	if repoSource.PushToRegistry.ID == "" {
		return nil
	}
	data.Step = stepImagePush

	e.addStepStartLog(ctx, data.taskData, "Start pushing image to registry...")
	defer e.addStepEndLog(ctx, data.taskData, timeutil.NowUTC(), err)

	regAuth := data.RefObjects.RefSettings[repoSource.PushToRegistry.ID]
	data.RegAuthHeader, err = regAuth.MustAsRegistryAuth().GenerateAuthHeader()
	if err != nil {
		return apperrors.Wrap(err)
	}

	for _, tag := range data.DeploymentOutput.ImageTags {
		if !strings.Contains(tag, "/") { // only push tag containing `/` in it
			continue
		}
		logsReader, err := e.dockerManager.ImagePush(ctx, tag, func(options *image.PushOptions) {
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
	}

	return nil
}

func (e *Executor) repoDeployStepServiceApply(
	ctx context.Context,
	data *repoDeployTaskData,
) (err error) {
	data.Step = stepServiceApply
	deployment := data.Deployment

	e.addStepStartLog(ctx, data.taskData, "Applying changes to service...")
	defer e.addStepEndLog(ctx, data.taskData, timeutil.NowUTC(), err)

	service, err := e.dockerManager.ServiceInspect(ctx, data.App.ServiceID)
	if err != nil {
		return apperrors.Wrap(err)
	}

	spec := &service.Spec
	contSpec := spec.TaskTemplate.ContainerSpec
	contSpec.Image = data.DeploymentOutput.ImageTags[0]
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

func (e *Executor) repoDeployStepPrepare(
	_ context.Context,
	data *repoDeployTaskData,
) (err error) {
	deployment := data.Deployment
	repoSource := deployment.Settings.RepoSource

	// Loads repo credentials (github app, git token, ssh key) if configured
	if repoSource.Credentials.ID != "" {
		data.CredSetting = data.RefObjects.RefSettings[repoSource.Credentials.ID]
	}

	// Creates checkout dir
	data.CheckoutPath, err = fileutil.CreateTempDir("", "*", 0)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Parse repo URL
	repoURLInfo, err := vcsurl.Parse(repoSource.RepoURL)
	if err != nil {
		return apperrors.Wrap(err)
	}
	data.RepoURLInfo = repoURLInfo

	return nil
}

//nolint:unparam
func (e *Executor) repoDeployStepCleanup(
	_ context.Context,
	data *repoDeployTaskData,
) (err error) {
	if data.CheckoutPath != "" {
		_ = os.RemoveAll(data.CheckoutPath)
	}

	return nil
}

func (e *Executor) onRepoDeployCommand(
	ctx context.Context,
	taskData *repoDeployTaskData,
	cmd base.TaskCommand,
	_ ...any,
) {
	if cmd == base.TaskCommandCancel && taskData.Step == stepImageBuild {
		err := e.dockerManager.ImageBuildCancel(ctx, taskData.Task.ID)
		if err != nil {
			_ = taskData.LogStore.Add(ctx, applog.NewErrFrame("failed to cancel image build: "+
				err.Error(), applog.TsNow))
		}
	}
}
