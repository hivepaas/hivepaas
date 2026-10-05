package appdeploymentserviceimpl

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appdeploymentservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagebuildservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/repocheckoutservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

// fakeImageBuildService builds on this node and keeps what it was asked to
// build.
type fakeImageBuildService struct {
	imagebuildservice.Service
	inputs *imagebuildservice.BuildInputs
	built  *imagebuildservice.ImageBuildReq
}

func (f *fakeImageBuildService) ResolveBuildInputs(
	_ context.Context, _ database.IDB, _ *imagebuildservice.ImageBuildReq,
) (*imagebuildservice.BuildInputs, error) {
	return f.inputs, nil
}

func (f *fakeImageBuildService) SelectBuildWorkerNode(
	_ context.Context, _ *entity.ImageBuildSettings,
) (imagebuildservice.BuildNodeResp, error) {
	return imagebuildservice.BuildNodeResp{Node: &swarm.Node{ID: "self"}, CurrentNodeID: "self"}, nil
}

func (f *fakeImageBuildService) ImageBuild(
	_ context.Context, _ database.IDB, req *imagebuildservice.ImageBuildReq,
) (*imagebuildservice.ImageBuildResp, error) {
	f.built = req
	return &imagebuildservice.ImageBuildResp{ImageTags: []string{"fn:dev-1234567"}}, nil
}

// fakeRepoCheckoutService checks out a repository holding two functions.
type fakeRepoCheckoutService struct {
	repocheckoutservice.Service
	req *repocheckoutservice.RepoCheckoutReq
}

func (f *fakeRepoCheckoutService) Checkout(
	_ context.Context, req *repocheckoutservice.RepoCheckoutReq,
) (*repocheckoutservice.RepoCheckoutResp, error) {
	f.req = req
	for _, name := range []string{"fns/hello/index.js", "fns/bye/index.js"} {
		path := filepath.Join(req.CheckoutDir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(path, []byte("export default () => {}"), 0o600); err != nil {
			return nil, err
		}
	}
	return &repocheckoutservice.RepoCheckoutResp{CommitHash: "0123456789abcdef", CommitTitle: "fix: hello"}, nil
}

func functionSource() *entity.DeploymentFunctionSource {
	return &entity.DeploymentFunctionSource{
		Runtime:        base.FunctionRuntimeNode24,
		Contract:       base.FunctionContractV1,
		Entrypoint:     entity.FunctionEntrypoint{File: "index.js", Handler: "default"},
		Timeout:        timeutil.Duration(30 * time.Second),
		MaxConcurrency: 16,
		MaxBodySize:    6 * unit.MB,
	}
}

func functionDeployData(t *testing.T, source *entity.DeploymentFunctionSource) *repoDeploymentData {
	t.Helper()
	config.SetCurrent(&config.Config{})
	t.Cleanup(func() { config.SetCurrent(nil) })
	tempDir := t.TempDir()
	return &repoDeploymentData{
		appDeploymentData: &appDeploymentData{
			AppDeploymentReq: &appdeploymentservice.AppDeploymentReq{
				TaskExecData: &queue.TaskExecData{
					Task: &entity.Task{ID: "task-1"}, LogStore: tasklog.NewNullStore(), RefObjects: entity.NewRefObjects(),
				},
			},
			App: &entity.App{ID: "app-1"},
			Deployment: &entity.Deployment{
				ID:       "deployment-1",
				Settings: &entity.AppDeploymentSettings{ActiveMethod: base.DeploymentMethodFunction, FunctionSource: source},
				Output:   &entity.AppDeploymentOutput{},
			},
			DeployArgs: &entity.TaskAppDeployArgs{},
		},
		TempDir:     tempDir,
		CheckoutDir: filepath.Join(tempDir, "checkout"),
	}
}

func inlineFunctionSource() *entity.DeploymentFunctionSource {
	source := functionSource()
	source.Code.Inline = &entity.FunctionInlineCode{Files: []*entity.FunctionFile{
		{Path: "index.js", Content: "export default () => {}"},
		{Path: "package.json", Content: `{"dependencies": {"ms": "2.1.3"}}`},
	}}
	return source
}

// A function's inline code is the source of its build: its files, and nothing
// else.
func TestAFunctionsInlineCodeIsTheSourceOfItsBuild(t *testing.T) {
	data := functionDeployData(t, inlineFunctionSource())

	err := (&service{}).functionDeployStepSource(context.Background(), data)

	assert.NoError(t, err)
	assert.Equal(t, data.CheckoutDir, data.ContextDir)
	content, err := os.ReadFile(filepath.Join(data.ContextDir, "package.json"))
	assert.NoError(t, err)
	assert.JSONEq(t, `{"dependencies": {"ms": "2.1.3"}}`, string(content))
}

// A function in a repository is built from its directory, at the commit the
// checkout found, which the deployment records.
func TestAFunctionInARepositoryIsBuiltFromItsDirectory(t *testing.T) {
	source := functionSource()
	source.Code = entity.FunctionCode{Dir: "fns/hello", Repo: &entity.FunctionRepoCode{
		RepoType: base.RepoTypeGit, RepoID: "github.com/acme/fns", RepoURL: "https://github.com/acme/fns.git",
		RepoRef: "refs/heads/main",
	}}
	data := functionDeployData(t, source)
	checkout := &fakeRepoCheckoutService{}

	err := (&service{repoCheckoutService: checkout}).functionDeployStepSource(context.Background(), data)

	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(data.CheckoutDir, "fns", "hello"), data.ContextDir)
	assert.Equal(t, "https://github.com/acme/fns.git", checkout.req.RepoSource.RepoURL)
	assert.Equal(t, "0123456789abcdef", source.Code.Repo.CommitHash)
	assert.Equal(t, "fix: hello", data.Deployment.Output.CommitTitle)
}

func TestAFunctionsDirectoryThatIsNotInItsRepositoryFailsTheDeployment(t *testing.T) {
	source := functionSource()
	source.Code = entity.FunctionCode{Dir: "fns/missing", Repo: &entity.FunctionRepoCode{
		RepoType: base.RepoTypeGit, RepoURL: "https://github.com/acme/fns.git",
	}}
	data := functionDeployData(t, source)
	var logged []string
	data.LogStore = tasklog.NewForwardStore("deployment-1", func(_ context.Context, frames []*tasklog.LogFrame) error {
		for _, frame := range frames {
			logged = append(logged, frame.Data)
		}
		return nil
	})

	err := (&service{repoCheckoutService: &fakeRepoCheckoutService{}}).functionDeployStepSource(
		context.Background(), data)

	assert.ErrorIs(t, err, hperrors.ErrNotFound)
	assert.Contains(t, strings.Join(logged, "\n"), "no directory 'fns/missing'")
}

// A function is built from the Dockerfile written for it, which mounts the
// build's secrets, and its image is tagged after its code.
func TestAFunctionIsBuiltFromTheDockerfileWrittenForIt(t *testing.T) {
	source := inlineFunctionSource()
	source.PushToRegistry = entity.ObjectID{ID: "registry-1"}
	data := functionDeployData(t, source)
	production := "production"
	builds := &fakeImageBuildService{inputs: &imagebuildservice.BuildInputs{
		EnvVars:       map[string]*string{"NODE_ENV": &production},
		SecretEnvVars: map[string]string{"NPM_TOKEN": "t0ken"},
		Secrets:       []string{"t0ken"},
	}}
	s := &service{imageBuildService: builds}
	assert.NoError(t, s.functionDeployStepSource(context.Background(), data))

	err := s.functionDeployStepImageBuild(context.Background(), database.Tx{}, data)

	assert.NoError(t, err)
	built := builds.built
	assert.Equal(t, base.DockerfileSourceManual, built.Dockerfile.Source)
	assert.Equal(t, ".hivepaas/Dockerfile", built.Dockerfile.Path)
	assert.Contains(t, built.Dockerfile.Content, "FROM "+base.StableVersion.FunctionRuntimes["node24"]+" AS base")
	assert.Contains(t, built.Dockerfile.Content, "ARG NODE_ENV\n")
	assert.Contains(t, built.Dockerfile.Content, "--mount=type=secret,id=NPM_TOKEN,env=NPM_TOKEN")
	assert.Equal(t, data.ContextDir, built.CheckoutDir)
	assert.Len(t, built.CommitHash, 64)
	assert.Equal(t, entity.ObjectID{ID: "registry-1"}, built.PushToRegistry)
	assert.Same(t, builds.inputs, built.Inputs)
	assert.Equal(t, []string{"fn:dev-1234567"}, data.Deployment.Output.ImageTags)
	assert.Equal(t, map[string]string{"node24": base.StableVersion.FunctionRuntimes["node24"]},
		data.Deployment.Output.RuntimeImages)
}
