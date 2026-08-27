package appdeploymentserviceimpl

import (
	"context"
	"fmt"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	imagebuildserviceclient "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/client/imagebuildservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagebuildservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/imagebuildagentuc/imagebuildagentdto"
)

const (
	buildWorkerNodeWaitTimeout  = 10 * time.Minute
	buildWorkerNodePollInterval = 3 * time.Second
)

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

	buildNode, err := s.repoDeployStepGetBuildNode(ctx, data)
	if err != nil {
		return hperrors.Wrap(err)
	}
	defer func() {
		buildNode.ReleaseNode()
	}()

	isRemote := buildNode.Node != nil && buildNode.Node.ID != "" && buildNode.Node.ID != buildNode.CurrentNodeID
	if config.Current.DevMode.Enabled && config.Current.DevMode.ForceAgentLocal {
		isRemote = true
	}

	for isRemote { // loop at most 1 time
		nodeID, nodeName := buildNode.CurrentNodeID, "<current>"
		if buildNode.Node != nil {
			nodeID = buildNode.Node.ID
			nodeName = gofn.Coalesce(buildNode.Node.Spec.Name, buildNode.Node.Description.Hostname)
		}
		agentAddr, err := s.agentService.GetAgentAddrForNode(ctx, nodeID)
		if err != nil {
			_ = data.LogStore.Add(ctx, tasklog.NewWarnFrame(
				fmt.Sprintf("Failed to get agent address for node '%s' (id '%s'): %v. "+
					"Falling back to current node.", nodeName, nodeID, err),
				tasklog.TsNow))
			buildNode.ReleaseNode()
			break
		}

		agentClient, err := imagebuildserviceclient.NewImageBuildServiceClient(agentAddr)
		if err != nil {
			_ = data.LogStore.Add(ctx, tasklog.NewWarnFrame(
				fmt.Sprintf("Failed to connect to agent on node '%s' (id '%s'): %v. "+
					"Falling back to current node.", nodeName, nodeID, err),
				tasklog.TsNow))
			buildNode.ReleaseNode()
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
				buildNode.ReleaseNode()
				break
			}
			return hperrors.Wrap(err)
		}

		data.Deployment.Output.ImageTags = buildResp.ImageTags
		return nil //nolint:staticcheck
	}

	buildResp, err := s.imageBuildService.ImageBuild(ctx, db, buildReq)
	if err != nil {
		return hperrors.Wrap(err)
	}

	data.Deployment.Output.ImageTags = buildResp.ImageTags

	return nil
}

func (s *service) repoDeployStepGetBuildNode(
	ctx context.Context,
	data *repoDeploymentData,
) (buildNode imagebuildservice.BuildNodeResp, err error) {
	buildNode, err = s.imageBuildService.SelectBuildWorkerNode(ctx, data.ImageBuildSettings)
	if err != nil {
		return buildNode, hperrors.Wrap(err)
	}
	defer func() {
		if err != nil {
			buildNode.ReleaseNode()
		}
	}()

	// There is a node available, return
	if buildNode.Node != nil {
		return buildNode, nil
	}

	// Wait for a node to be available

	_ = data.LogStore.Add(ctx, tasklog.NewWarnFrame(
		fmt.Sprintf("All build nodes are currently at maximum capacity. "+
			"Waiting for an available build slot (timeout: %s)...", buildWorkerNodeWaitTimeout),
		tasklog.TsNow))

	waitTimer := time.NewTimer(buildWorkerNodeWaitTimeout)
	defer waitTimer.Stop()

	pollTicker := time.NewTicker(buildWorkerNodePollInterval)
	defer pollTicker.Stop()

	startTime := time.Now()
	lastLogTime := startTime

	for {
		select {
		case <-ctx.Done():
			return buildNode, hperrors.Wrap(ctx.Err())

		case <-waitTimer.C:
			return buildNode, hperrors.NewUnavailable("Build worker node (timed out waiting for available slot)")

		case <-pollTicker.C:
			buildNode, err = s.imageBuildService.SelectBuildWorkerNode(ctx, data.ImageBuildSettings)
			if err != nil {
				return buildNode, hperrors.Wrap(err)
			}
			if buildNode.Node != nil {
				_ = data.LogStore.Add(ctx, tasklog.NewOutFrame(
					fmt.Sprintf("Build slot acquired after %s. Continuing build...",
						time.Since(startTime).Truncate(time.Second)), tasklog.TsNow))
				return buildNode, nil
			}

			if time.Since(lastLogTime) >= 30*time.Second { //nolint:mnd
				lastLogTime = time.Now()
				_ = data.LogStore.Add(ctx, tasklog.NewWarnFrame(
					fmt.Sprintf("Still waiting for an available build slot (elapsed: %s)...",
						time.Since(startTime).Truncate(time.Second)), tasklog.TsNow))
			}
		}
	}
}
