package appdeploymentserviceimpl

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	imagebuildserviceclient "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/client/imagebuildservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/dockerhelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/fileutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagebuildservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/repocheckoutservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/imagebuildagentuc/imagebuildagentdto"
)

const (
	stepRepoCheckout = "repo-checkout"
	stepImageBuild   = "image-build"
	stepServiceApply = "service-apply"

	dockerServiceApplyRetryMax = 2
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
	err = s.repoDeployStepServiceApply(ctx, data)
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

func (s *service) repoDeployStepSourceCheckout(
	ctx context.Context,
	data *repoDeploymentData,
) (err error) {
	data.Step = stepRepoCheckout
	deployment := data.Deployment
	repoSource := deployment.Settings.RepoSource

	checkoutReq := &repocheckoutservice.RepoCheckoutReq{
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
		return apperrors.Wrap(err)
	}

	repoSource.CommitHash = checkoutResp.CommitHash
	data.Deployment.Output.CommitHash = checkoutResp.CommitHash
	data.Deployment.Output.CommitMessage = checkoutResp.CommitMessage
	data.Deployment.Output.CommitTitle = checkoutResp.CommitTitle
	data.Deployment.Output.CommitAuthor = checkoutResp.CommitAuthor

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

	s.addStepStartLog(ctx, data.appDeploymentData, "Start building docker image...")
	defer s.addStepEndLog(ctx, data.appDeploymentData, timeutil.NowUTC(), err)

	buildReq := &imagebuildservice.ImageBuildReq{
		TaskExecData: &queue.TaskExecData{
			Task:       data.Task,
			RefObjects: data.RefObjects,
			LogStore:   data.LogStore,
		},
		App:                data.App,
		CommitHash:         repoSource.CommitHash,
		Dockerfile:         repoSource.Dockerfile,
		ImageName:          repoSource.ImageName,
		PushToRegistry:     repoSource.PushToRegistry,
		ImageBuildSettings: data.ImageBuildSettings,
		BuildID:            data.Task.ID,
		CheckoutDir:        data.CheckoutDir,
	}
	if deployment.Settings.NoCache || (data.ImageBuildSettings != nil && data.ImageBuildSettings.NoCache) {
		buildReq.NoCache = true
	}

	var buildWorkerNode *swarm.Node
	if data.ImageBuildSettings != nil {
		buildWorkerNode, err = s.imageBuildService.SelectBuildWorkerNode(ctx, data.ImageBuildSettings)
		if err != nil {
			return apperrors.Wrap(err)
		}
	}

	currNodeID, err := s.dockerManager.NodeCurrentID(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}

	isRemote := buildWorkerNode != nil && buildWorkerNode.ID != "" && buildWorkerNode.ID != currNodeID
	if config.Current.DevMode.Enabled && config.Current.DevMode.ForceAgentLocal {
		isRemote = true
	}

	for isRemote { // loop at most 1 time
		nodeID, nodeName := currNodeID, "<current>"
		if buildWorkerNode != nil {
			nodeID = buildWorkerNode.ID
			nodeName = gofn.Coalesce(buildWorkerNode.Spec.Name, buildWorkerNode.Description.Hostname)
		}
		agentAddr, err := s.agentService.GetAgentAddrForNode(ctx, nodeID)
		if err != nil {
			_ = data.LogStore.Add(ctx, tasklog.NewWarnFrame(
				fmt.Sprintf("Failed to get agent address for node '%s' (id '%s'): %v. "+
					"Falling back to current node.", nodeName, nodeID, err),
				tasklog.TsNow))
			break
		}

		agentClient, err := imagebuildserviceclient.NewImageBuildServiceClient(agentAddr)
		if err != nil {
			_ = data.LogStore.Add(ctx, tasklog.NewWarnFrame(
				fmt.Sprintf("Failed to connect to agent on node '%s' (id '%s'): %v. "+
					"Falling back to current node.", nodeName, nodeID, err),
				tasklog.TsNow))
			break
		}
		defer agentClient.Close()

		_ = data.LogStore.Add(ctx, tasklog.NewOutFrame(
			fmt.Sprintf("Starting build process on worker node '%s' (id '%s')...", nodeName, nodeID),
			tasklog.TsNow))

		agentReq := &imagebuildagentdto.ImageBuildReq{
			TaskID:        data.Task.ID,
			AppID:         data.App.ID,
			ImageBuildReq: *buildReq,
			SendLog: func(frames []*tasklog.LogFrame) error {
				return data.LogStore.Add(ctx, frames...)
			},
		}

		timeStart := time.Now()
		buildResp, err := agentClient.ImageBuild(ctx, agentReq)
		if err != nil {
			if time.Since(timeStart) < 10*time.Second && ctx.Err() == nil { //nolint:mnd
				// If any error occurs within 10s, we can fall back to the current node.
				// Otherwise, it is likely an error in the source code where retrying will make no difference.
				_ = data.LogStore.Add(ctx, tasklog.NewWarnFrame(
					fmt.Sprintf("Failed to build image via agent on node '%s' (id '%s'): %v. "+
						"Falling back to current node.", nodeName, nodeID, err),
					tasklog.TsNow))
				break
			}
			return apperrors.Wrap(err)
		}

		data.Deployment.Output.ImageTags = buildResp.ImageTags
		return nil //nolint:staticcheck
	}

	buildResp, err := s.imageBuildService.ImageBuild(ctx, db, buildReq)
	if err != nil {
		return apperrors.Wrap(err)
	}

	data.Deployment.Output.ImageTags = buildResp.ImageTags

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
			return apperrors.Wrap(err)
		}
	}

	queryRegistry := false
	for i := range dockerServiceApplyRetryMax + 1 {
		if i > 0 {
			queryRegistry = true
			select {
			case <-ctx.Done():
				err = apperrors.Wrap(ctx.Err())
				break
			case <-time.After(time.Duration(1+i) * time.Second):
			}
		}

		inspect, e := s.dockerManager.ServiceInspect(ctx, data.App.ServiceID)
		if e != nil { // error, need to retry
			if errors.Is(e, apperrors.ErrNotFound) {
				err = apperrors.Wrap(e)
				break
			}
			err = apperrors.Wrap(e)
			continue
		}
		service := &inspect.Service
		spec := &service.Spec
		contSpec := spec.TaskTemplate.ContainerSpec
		contSpec.Image = data.Deployment.Output.ImageTags[0]
		contSpec.Dir = deployment.Settings.WorkingDir
		dockerhelper.ContainerCommandApply(contSpec, deployment.Settings.Command)

		_, e = s.dockerManager.ServiceUpdate(ctx, data.App.ServiceID, &service.Version, spec,
			func(options *client.ServiceUpdateOptions) {
				options.EncodedRegistryAuth = regAuthHeader
				options.QueryRegistry = queryRegistry
			})
		if e != nil { // error, need to retry
			err = apperrors.Wrap(e)
			continue
		}
		// successful, no need to retry
		err = nil
		break
	}
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
