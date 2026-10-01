package functionserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	fnservice "github.com/hivepaas/hivepaas/hivepaas_app/service/functionservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/functionservice/functiontest"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagebuildservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/systemappservice"
)

func (s *service) TestRun(
	ctx context.Context,
	db database.IDB,
	req *fnservice.TestRunReq,
) (*functiontest.RunResp, error) {
	runReq, buildSettings, err := s.runReq(ctx, db, req)
	if err != nil {
		return nil, err
	}

	// A build node, as for a build: the run installs the libraries there, and
	// its slot is given back when the run ends.
	node, err := s.imageBuildService.SelectBuildWorkerNode(ctx, buildSettings)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	defer node.ReleaseNode()
	if node.Node == nil {
		return nil, hperrors.NewUnavailable("Build worker node (every one is at its maximum of builds)")
	}

	remote := node.Node.ID != "" && node.Node.ID != node.CurrentNodeID
	if config.Current().DevMode.Enabled && config.Current().DevMode.ForceAgentLocal {
		remote = true
	}
	if !remote {
		resp, err := s.localRunner.Run(ctx, runReq)
		return resp, hperrors.Wrap(err)
	}

	addr, err := s.agentService.GetAgentAddrForNode(ctx, node.Node.ID)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	agent, err := s.newAgentClient(addr)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	defer agent.Close()
	resp, err := agent.TestRun(ctx, runReq)
	return resp, hperrors.Wrap(err)
}

// runReq is the run of the function as it is deployed: its build's inputs for
// the install, its environment, its resources and its network as its service
// has them, on the runtime images of the release running now.
func (s *service) runReq(
	ctx context.Context,
	db database.IDB,
	req *fnservice.TestRunReq,
) (*functiontest.RunReq, *entity.ImageBuildSettings, error) {
	app := req.App
	inputs, err := s.imageBuildService.ResolveBuildInputs(ctx, db, &imagebuildservice.ImageBuildReq{App: app})
	if err != nil {
		return nil, nil, hperrors.Wrap(err)
	}
	buildSettings, err := s.loadBuildSettings(ctx, db, app)
	if err != nil {
		return nil, nil, err
	}
	svc, err := s.clusterService.ServiceInspect(ctx, app.ServiceID, false)
	if err != nil {
		return nil, nil, hperrors.Wrap(err)
	}

	runReq := &functiontest.RunReq{
		Source:        req.Source,
		Files:         req.Files,
		Request:       req.Request,
		Images:        systemappservice.CurrentRelease().FunctionRuntimes,
		Inputs:        inputs,
		BuildSettings: buildSettings,
		Network:       s.networkService.GetProjectNetworkName(app.Project, app.ProjectEnv.Name),
	}
	task := svc.Spec.TaskTemplate
	if task.ContainerSpec != nil {
		runReq.Env = task.ContainerSpec.Env
	}
	if task.Resources != nil && task.Resources.Limits != nil {
		runReq.NanoCPUs = task.Resources.Limits.NanoCPUs
		runReq.MemoryBytes = task.Resources.Limits.MemoryBytes
	}
	return runReq, buildSettings, nil
}
