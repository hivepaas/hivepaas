package functiontest

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/functionservice/functionbuild"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagebuildservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

// defaultBuilder builds with docker's default builder: a test does not create
// the builder HivePaaS's own builds use.
type defaultBuilder struct{ t *testing.T }

func (b defaultBuilder) BuildLibraries(ctx context.Context, req *LibrariesBuildReq) (string, error) {
	path := filepath.Join(req.ContextDir, functionbuild.DockerfilePath)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, []byte(req.Dockerfile), 0o600); err != nil {
		return "", err
	}
	b.t.Cleanup(func() { _ = exec.Command("docker", "rmi", "-f", req.Image).Run() })
	build := exec.CommandContext(ctx, "docker", "buildx", "build", "--load", "-f", functionbuild.DockerfilePath,
		"-t", req.Image, ".")
	build.Dir = req.ContextDir
	out, err := build.CombinedOutput()
	return string(out), err
}

var realTestRuns = map[base.FunctionRuntime]struct {
	entrypoint entity.FunctionEntrypoint
	files      []*entity.FunctionFile
	lockFile   string
}{
	base.FunctionRuntimeNode24: {
		entity.FunctionEntrypoint{File: "index.js", Handler: "default"},
		[]*entity.FunctionFile{
			{Path: "package.json", Content: `{"type": "module", "dependencies": {"ms": "2.1.3"}}`},
			{Path: "index.js", Content: `import ms from 'ms'
export default (req) => {
  console.log('called with', req.text())
  return { body: { got: req.text(), library: String(ms('1m')), env: process.env.GREETING } }
}
`},
		},
		"package-lock.json",
	},
	base.FunctionRuntimeBun1: {
		entity.FunctionEntrypoint{File: "index.ts", Handler: "default"},
		[]*entity.FunctionFile{
			{Path: "package.json", Content: `{"type": "module", "dependencies": {"ms": "2.1.3"}}`},
			{Path: "index.ts", Content: `import ms from 'ms'
export default (req: { text(): string }) => {
  console.log('called with', req.text())
  return { body: { got: req.text(), library: String(ms('1m')), env: process.env.GREETING } }
}
`},
		},
		"bun.lock",
	},
	base.FunctionRuntimePython313: {
		entity.FunctionEntrypoint{File: "main.py", Handler: "handler"},
		[]*entity.FunctionFile{
			{Path: "requirements.txt", Content: "six==1.17.0\n"},
			{Path: "main.py", Content: `import os

import six


def handler(req, ctx):
    print("called with", req.text())
    return {"body": {"got": req.text(), "library": six.__version__, "env": os.environ.get("GREETING")}}
`},
		},
		"requirements.lock",
	},
	base.FunctionRuntimeGo127: {
		entity.FunctionEntrypoint{File: ".", Handler: "Handle"},
		[]*entity.FunctionFile{
			{Path: "go.mod", Content: "module example.com/fn\n\ngo 1.27\n"},
			{Path: "fn.go", Content: `package fn

import (
	"context"
	"os"

	"github.com/google/uuid"
	"github.com/hivepaas/function-runtimes/hivepaas"
)

func Handle(ctx context.Context, req *hivepaas.Request) (*hivepaas.Response, error) {
	hivepaas.Log(ctx, "called with %s", req.Text())
	return hivepaas.JSON(200, map[string]string{
		"got":     req.Text(),
		"library": uuid.NewSHA1(uuid.NameSpaceURL, []byte("hivepaas")).String()[:8],
		"env":     os.Getenv("GREETING"),
	})
}
`},
		},
		"go.sum",
	},
}

// End to end, with docker on this machine and the release's runtime images: a
// function of each runtime is test-run twice, its libraries installed by the
// first run only, its environment given, its lock file returned. It pulls the
// images and the libraries, so it runs only when HIVEPAAS_DOCKER_TESTS=1.
func TestAFunctionOfEachRuntimeIsTestRun(t *testing.T) {
	if os.Getenv("HIVEPAAS_DOCKER_TESTS") != "1" {
		t.Skip("set HIVEPAAS_DOCKER_TESTS=1 to test-run functions with docker")
	}
	manager, err := docker.New()
	if !assert.NoError(t, err) {
		return
	}
	runner := &Runner{Docker: manager, Builder: defaultBuilder{t: t}, TempDir: t.TempDir()}

	for runtime, fn := range realTestRuns {
		t.Run(string(runtime), func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
			defer cancel()
			req := &RunReq{
				Source: &entity.DeploymentFunctionSource{
					Runtime: runtime, Contract: base.FunctionContractV1, Entrypoint: fn.entrypoint,
					Timeout: timeutil.Duration(10 * time.Second), MaxConcurrency: 1, MaxBodySize: unit.MB,
				},
				Files:   fn.files,
				Request: &Request{Method: "POST", Path: "/", Body: []byte("Ada")},
				Images:  base.BetaVersion.FunctionRuntimes,
				Inputs:  &imagebuildservice.BuildInputs{},
				Env:     []string{"GREETING=hello"},
			}

			first, err := runner.Run(ctx, req)
			if !assert.NoError(t, err) {
				return
			}
			assert.Equal(t, OutcomeOK, first.Outcome, first.Error+first.LibrariesLog)
			assert.Equal(t, 200, first.Status)
			var body map[string]string
			assert.NoError(t, json.Unmarshal(first.Body, &body), string(first.Body))
			assert.Equal(t, "Ada", body["got"])
			assert.Equal(t, "hello", body["env"])
			assert.Contains(t, first.Logs, "called with")
			assert.True(t, first.LibrariesBuilt)
			found := false
			for _, lock := range first.LockFiles {
				found = found || lock.Path == fn.lockFile
			}
			assert.True(t, found, "%s comes back", fn.lockFile)

			second, err := runner.Run(ctx, req)
			assert.NoError(t, err)
			assert.Equal(t, OutcomeOK, second.Outcome, second.Error)
			assert.False(t, second.LibrariesBuilt)
		})
	}
}
