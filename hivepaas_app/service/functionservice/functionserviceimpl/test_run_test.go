package functionserviceimpl

import (
	"context"
	"testing"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/client/functionservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/agentservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clusterservice"
	fnservice "github.com/hivepaas/hivepaas/hivepaas_app/service/functionservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/functionservice/functiontest"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagebuildservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/networkservice"
)

// fakeBuilds resolves a build's inputs and picks a build node.
type fakeBuilds struct {
	imagebuildservice.Service
	node     string
	released bool
}

func (f *fakeBuilds) ResolveBuildInputs(
	_ context.Context, _ database.IDB, req *imagebuildservice.ImageBuildReq,
) (*imagebuildservice.BuildInputs, error) {
	return &imagebuildservice.BuildInputs{SecretEnvVars: map[string]string{"NPM_TOKEN": "for " + req.App.ID}}, nil
}

func (f *fakeBuilds) SelectBuildWorkerNode(
	_ context.Context, _ *entity.ImageBuildSettings,
) (imagebuildservice.BuildNodeResp, error) {
	return imagebuildservice.BuildNodeResp{
		Node: &swarm.Node{ID: f.node}, CurrentNodeID: "self",
		ReleaseNodeFunc: func() { f.released = true },
	}, nil
}

// fakeCluster has the function's service.
type fakeCluster struct {
	clusterservice.Service
}

func (fakeCluster) ServiceInspect(_ context.Context, serviceID string, _ bool) (*swarm.Service, error) {
	return &swarm.Service{ID: serviceID, Spec: swarm.ServiceSpec{TaskTemplate: swarm.TaskSpec{
		ContainerSpec: &swarm.ContainerSpec{Env: []string{"GREETING=hello"}},
		Resources: &swarm.ResourceRequirements{Limits: &swarm.Limit{
			NanoCPUs: 250_000_000, MemoryBytes: 128 << 20,
		}},
	}}}, nil
}

type fakeNetworks struct {
	networkservice.Service
}

func (fakeNetworks) GetProjectNetworkName(project *entity.Project, env string) string {
	return project.Key + "_" + env + "_net"
}

type fakeAgents struct {
	agentservice.Service
}

func (fakeAgents) GetAgentAddrForNode(_ context.Context, nodeID string) (string, error) {
	return nodeID + ":50051", nil
}

// fakeRunner is this node's runner, or an agent's.
type fakeRunner struct {
	got  *functiontest.RunReq
	addr string
}

func (f *fakeRunner) Run(_ context.Context, req *functiontest.RunReq) (*functiontest.RunResp, error) {
	f.got = req
	return &functiontest.RunResp{Outcome: functiontest.OutcomeOK, Status: 200}, nil
}

func (f *fakeRunner) TestRun(ctx context.Context, req *functiontest.RunReq) (*functiontest.RunResp, error) {
	return f.Run(ctx, req)
}

func (f *fakeRunner) Close() error { return nil }

func testService(t *testing.T, node string) (*service, *fakeBuilds, *fakeRunner, *fakeRunner) {
	t.Helper()
	config.SetCurrent(&config.Config{})
	t.Cleanup(func() { config.SetCurrent(nil) })
	builds := &fakeBuilds{node: node}
	local, remote := &fakeRunner{}, &fakeRunner{}
	s := &service{
		imageBuildService: builds,
		clusterService:    fakeCluster{},
		networkService:    fakeNetworks{},
		agentService:      fakeAgents{},
		localRunner:       local,
		newAgentClient: func(addr string) (functionservice.FunctionServiceClient, error) {
			remote.addr = addr
			return remote, nil
		},
		loadBuildSettings: func(context.Context, database.IDB, *entity.App) (*entity.ImageBuildSettings, error) {
			return &entity.ImageBuildSettings{NoCache: true}, nil
		},
	}
	return s, builds, local, remote
}

func testRunReq() *fnservice.TestRunReq {
	return &fnservice.TestRunReq{
		App: &entity.App{ID: "app-1", ServiceID: "svc-1", Project: &entity.Project{Key: "shop"},
			ProjectEnv: &entity.ProjectEnv{Name: "dev"}},
		Source: &entity.DeploymentFunctionSource{Runtime: base.FunctionRuntimeNode24,
			Timeout: timeutil.Duration(30 * time.Second)},
		Files:   []*entity.FunctionFile{{Path: "index.js", Content: "export default () => {}"}},
		Request: &functiontest.Request{Method: "GET", Path: "/"},
	}
}

// A test run is the function's: its build's inputs, its environment, its
// resources and network as its service has them, the release's runtime
// images, the code and the request it was given.
func TestATestRunIsTheFunctions(t *testing.T) {
	s, builds, local, _ := testService(t, "self")

	resp, err := s.TestRun(context.Background(), nil, testRunReq())

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.Status)
	got := local.got
	assert.Equal(t, "for app-1", got.Inputs.SecretEnvVars["NPM_TOKEN"])
	assert.Equal(t, []string{"GREETING=hello"}, got.Env)
	assert.Equal(t, int64(250_000_000), got.NanoCPUs)
	assert.Equal(t, int64(128<<20), got.MemoryBytes)
	assert.Equal(t, "shop_dev_net", got.Network)
	assert.Equal(t, base.StableVersion.FunctionRuntimes, got.Images)
	assert.Equal(t, "index.js", got.Files[0].Path)
	assert.Equal(t, "/", got.Request.Path)
	assert.True(t, got.BuildSettings.NoCache)
	assert.True(t, builds.released, "the build node's slot is given back")
}

// On a build node other than this one, the run goes to that node's agent.
func TestATestRunOnAnotherNodeGoesToItsAgent(t *testing.T) {
	s, builds, local, remote := testService(t, "node-2")

	_, err := s.TestRun(context.Background(), nil, testRunReq())

	assert.NoError(t, err)
	assert.Nil(t, local.got)
	assert.Equal(t, "node-2:50051", remote.addr)
	assert.Equal(t, []string{"GREETING=hello"}, remote.got.Env)
	assert.True(t, builds.released)
}
