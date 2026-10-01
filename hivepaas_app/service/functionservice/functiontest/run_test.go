package functiontest

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagebuildservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

var testImages = map[string]string{
	"node24":      "ghcr.io/hivepaas/function-runtime-node24:1.0.0@sha256:node",
	"python313":   "ghcr.io/hivepaas/function-runtime-python313:1.0.0@sha256:python",
	"go127":       "ghcr.io/hivepaas/function-runtime-go127:1.0.0@sha256:go",
	"go127-build": "ghcr.io/hivepaas/function-runtime-go127-build:1.0.0@sha256:gobuild",
}

func nodeSource() *entity.DeploymentFunctionSource {
	return &entity.DeploymentFunctionSource{
		Runtime:        base.FunctionRuntimeNode24,
		Contract:       base.FunctionContractV1,
		Entrypoint:     entity.FunctionEntrypoint{File: "index.js", Handler: "default"},
		Timeout:        timeutil.Duration(30 * time.Second),
		MaxConcurrency: 16,
		MaxBodySize:    6 * unit.MB,
	}
}

func runReq(files ...*entity.FunctionFile) *RunReq {
	if len(files) == 0 {
		files = []*entity.FunctionFile{
			{Path: "index.js", Content: "export default () => ({ body: 'hi' })"},
			{Path: "package.json", Content: `{"dependencies": {"ms": "2.1.3"}}`},
			{Path: "lib/util.js", Content: "export const x = 1"},
		}
	}
	return &RunReq{
		Source:      nodeSource(),
		Files:       files,
		Request:     &Request{Method: "POST", Path: "/hello", Body: []byte(`{"name":"Ada"}`)},
		Images:      testImages,
		Inputs:      &imagebuildservice.BuildInputs{SecretEnvVars: map[string]string{"NPM_TOKEN": "t0ken"}},
		Env:         []string{"DATABASE_URL=postgres://db/app", "HP_FN_TIMEOUT_MS=1"},
		Network:     "shop_dev_net",
		NanoCPUs:    500_000_000,
		MemoryBytes: 256 << 20,
	}
}

// result is invoke's last line, for a call answered 200.
func result(t *testing.T, body string) string {
	t.Helper()
	line, err := json.Marshal(map[string]any{
		"status": 200, "headers": map[string][]string{"content-type": {"text/plain"}},
		"body": base64.StdEncoding.EncodeToString([]byte(body)), "requestId": "req-1",
		"durationMs": 3.5, "outcome": "ok",
	})
	assert.NoError(t, err)
	return ResultMarker + string(line) + "\n"
}

func runner(t *testing.T) (*Runner, *fakeDocker, *fakeBuilder) {
	t.Helper()
	d := newFakeDocker(t)
	b := &fakeBuilder{docker: d}
	return &Runner{Docker: d, Builder: b, TempDir: t.TempDir()}, d, b
}

// A test run calls the handler once in a container of the function's own: its
// libraries' image, its environment with the limits it has, its network, its
// resources, and no published port. The container is removed afterwards.
func TestATestRunCallsTheHandlerInAContainerOfItsOwn(t *testing.T) {
	r, d, _ := runner(t)
	d.stdout = `{"hp":"log","msg":"hello"}` + "\n" + result(t, "hi Ada")

	resp, err := r.Run(context.Background(), runReq())

	assert.NoError(t, err)
	assert.Equal(t, OutcomeOK, resp.Outcome)
	assert.Equal(t, 200, resp.Status)
	assert.Equal(t, "hi Ada", string(resp.Body))
	assert.Equal(t, []string{"text/plain"}, resp.Headers["content-type"])
	assert.Equal(t, "req-1", resp.RequestID)
	assert.InDelta(t, 3.5, resp.DurationMs, 0.001)
	assert.Equal(t, `{"hp":"log","msg":"hello"}`+"\n", resp.Logs)

	created := d.created
	assert.True(t, strings.HasPrefix(created.Name, docker.TempContainerPrefix), created.Name)
	assert.Equal(t, docker.LabelTempResourceVal, created.Config.Labels[docker.LabelTempResource])
	assert.True(t, strings.HasPrefix(created.Config.Image, "hivepaas-function-libs:"), created.Config.Image)
	assert.Equal(t, []string{"sh", "-c", "exec hivepaas-runtime invoke < " + requestPath}, created.Config.Cmd)
	env := created.Config.Env
	assert.Contains(t, env, "DATABASE_URL=postgres://db/app")
	assert.Contains(t, env, "HP_FN_ENTRYPOINT=index.js")
	assert.Contains(t, env, "HP_FN_HANDLER=default")
	assert.Contains(t, env, "HP_FN_TIMEOUT_MS=30000")
	assert.NotContains(t, env, "HP_FN_TIMEOUT_MS=1", "the function's limits, not what its environment says")
	assert.Equal(t, container.NetworkMode("shop_dev_net"), created.HostConfig.NetworkMode)
	assert.Empty(t, created.HostConfig.PortBindings)
	assert.False(t, created.HostConfig.PublishAllPorts)
	assert.Equal(t, int64(500_000_000), created.HostConfig.NanoCPUs)
	assert.Equal(t, int64(256<<20), created.HostConfig.Memory)
	assert.True(t, d.started)
	assert.True(t, d.removed)
}

// The code is copied into /app, owned by the runtime's user, which writes there:
// npm and go mod tidy do. The request goes in a file the command reads.
func TestTheCodeIsTheRuntimeUsers(t *testing.T) {
	r, d, _ := runner(t)
	d.stdout = result(t, "")

	_, err := r.Run(context.Background(), runReq())

	assert.NoError(t, err)
	for _, path := range []string{"/app/index.js", "/app/package.json", "/app/lib/util.js"} {
		file, ok := d.copied[path]
		if assert.True(t, ok, path) {
			assert.Equal(t, RuntimeUID, file.uid, path)
		}
	}
	assert.Equal(t, "export const x = 1", d.copied["/app/lib/util.js"].content)
	request := d.copied["/tmp/hivepaas-request.json"]
	assert.Equal(t, RuntimeUID, request.uid)
	var sent map[string]any
	assert.NoError(t, json.Unmarshal([]byte(request.content), &sent))
	assert.Equal(t, "POST", sent["method"])
	assert.Equal(t, base64.StdEncoding.EncodeToString([]byte(`{"name":"Ada"}`)), sent["body"])
}

// The libraries' image is built once per node: a second run with the same
// manifest finds it, and the build is told its name, its Dockerfile and the
// build's inputs.
func TestTheLibrariesAreBuiltOnce(t *testing.T) {
	r, d, b := runner(t)
	d.stdout = result(t, "")

	first, err := r.Run(context.Background(), runReq())
	assert.NoError(t, err)
	assert.True(t, first.LibrariesBuilt)
	assert.Equal(t, "#5 DONE 1.2s", first.LibrariesLog)
	second, err := r.Run(context.Background(), runReq())
	assert.NoError(t, err)
	assert.False(t, second.LibrariesBuilt)
	assert.Empty(t, second.LibrariesLog)

	if assert.Len(t, b.built, 1) {
		built := b.built[0]
		assert.Equal(t, d.created.Config.Image, built.Image)
		assert.Contains(t, built.Dockerfile, "RUN --mount=type=secret,id=NPM_TOKEN,env=NPM_TOKEN")
		assert.Equal(t, "t0ken", built.Inputs.SecretEnvVars["NPM_TOKEN"])
	}
}

// A function with nothing to install runs on its runtime's image, pulled when
// the node does not have it.
func TestAFunctionWithoutLibrariesRunsOnItsRuntimesImage(t *testing.T) {
	r, d, b := runner(t)
	d.stdout = result(t, "")

	only := &entity.FunctionFile{Path: "index.js", Content: "export default () => {}"}
	_, err := r.Run(context.Background(), runReq(only))

	assert.NoError(t, err)
	assert.Empty(t, b.built)
	assert.Equal(t, []string{testImages["node24"]}, d.pulled)
	assert.Equal(t, testImages["node24"], d.created.Config.Image)
}

// Libraries that do not install end the run with the install's log, before a
// container exists.
func TestLibrariesThatDoNotInstallEndTheRun(t *testing.T) {
	r, d, b := runner(t)
	b.fail = true

	resp, err := r.Run(context.Background(), runReq())

	assert.NoError(t, err)
	assert.Equal(t, OutcomeLibrariesFailed, resp.Outcome)
	assert.Contains(t, resp.LibrariesLog, "404 Not Found")
	assert.Nil(t, d.created)
}

// A handler that cannot be loaded writes no result: the run says so, with what
// the runtime wrote on its standard error.
func TestAHandlerThatCannotBeLoadedSaysWhy(t *testing.T) {
	r, d, _ := runner(t)
	d.exitCode = 3
	d.stderr = "hivepaas: the handler could not be loaded: SyntaxError: Unexpected token\n"

	resp, err := r.Run(context.Background(), runReq())

	assert.NoError(t, err)
	assert.Equal(t, OutcomeNotLoaded, resp.Outcome)
	assert.Contains(t, resp.Error, "SyntaxError")
	assert.Zero(t, resp.Status)
	assert.True(t, d.removed)
}

// A container that outlives the call's timeout and the margin after it is
// killed, and still removed.
func TestAContainerThatOutlivesItsTimeoutIsKilled(t *testing.T) {
	r, d, _ := runner(t)
	r.KillMargin = 10 * time.Millisecond
	d.never = true
	req := runReq()
	req.Source.Timeout = timeutil.Duration(10 * time.Millisecond)

	resp, err := r.Run(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, OutcomeKilled, resp.Outcome)
	assert.True(t, d.killed)
	assert.True(t, d.removed)
}

// What comes back is cut where it would not fit a screen: the body and the
// logs, at 1 MB each, saying so.
func TestABodyAndLogsAreCutAt1MB(t *testing.T) {
	r, d, _ := runner(t)
	big := strings.Repeat("x", int(unit.MB)+10)
	d.stdout = big + "\n" + result(t, big)

	resp, err := r.Run(context.Background(), runReq())

	assert.NoError(t, err)
	assert.Len(t, resp.Body, int(unit.MB))
	assert.True(t, resp.BodyTruncated)
	assert.Len(t, resp.Logs, int(unit.MB))
	assert.True(t, resp.LogsTruncated)
}

// The lock file the install made, or a Go build tidied, comes back for the
// editor to add to the code; one the code already had, unchanged, does not.
func TestTheLockFileARunMadeComesBack(t *testing.T) {
	r, d, _ := runner(t)
	d.stdout = result(t, "")
	d.files["package-lock.json"] = `{"lockfileVersion": 3}`

	resp, err := r.Run(context.Background(), runReq())
	assert.NoError(t, err)
	assert.Equal(t, []*entity.FunctionFile{{Path: "package-lock.json", Content: `{"lockfileVersion": 3}`}},
		resp.LockFiles)

	req := runReq()
	req.Files = append(req.Files, &entity.FunctionFile{Path: "package-lock.json", Content: `{"lockfileVersion": 3}`})
	resp, err = r.Run(context.Background(), req)
	assert.NoError(t, err)
	assert.Empty(t, resp.LockFiles)
}

// The environment a test run gives the function is its own, the limits last:
// whatever the environment says about them, the function's settings win.
func TestTheLimitsAreTheFunctionsSettings(t *testing.T) {
	env := containerEnv(runReq())

	assert.Equal(t, "HP_FN_MAX_BODY_SIZE=6291456", env[len(env)-1])
	assert.Equal(t, 1, len(slices.DeleteFunc(slices.Clone(env), func(v string) bool {
		return !strings.HasPrefix(v, "HP_FN_TIMEOUT_MS=")
	})))
}

// A container that ends without a result - out of memory, say - says how it
// exited.
func TestARunWithoutAResultSaysHowItEnded(t *testing.T) {
	r, d, _ := runner(t)
	d.exitCode = 137

	resp, err := r.Run(context.Background(), runReq())

	assert.NoError(t, err)
	assert.Equal(t, OutcomeNoResult, resp.Outcome)
	assert.Equal(t, int64(137), resp.ExitCode)
}
