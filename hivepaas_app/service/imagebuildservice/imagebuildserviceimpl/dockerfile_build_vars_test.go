package imagebuildserviceimpl

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagebuildservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

const generatedDockerfile = `ARG NODE_VERSION=24
FROM node:${NODE_VERSION}-slim AS deps
# Install
RUN --mount=type=cache,target=/root/.npm \
    npm ci

from node:24-slim as runtime
run echo RUN inside a continued line \
    RUN not an instruction
CMD ["node", "server.js"]`

// Each stage declares the build's plain variables after its FROM, and each RUN
// mounts the secret ones: a continued line, a comment and the CMD stay as they
// were.
func TestBuildVariablesReachEveryStageAndStep(t *testing.T) {
	apiURL := "https://api.example.com"
	got := withBuildVariables(generatedDockerfile,
		map[string]*string{"VITE_API_URL": &apiURL, "B_FLAG": nil},
		map[string]string{"NPM_TOKEN": "s3cr3t"})

	want := `ARG NODE_VERSION=24
FROM node:${NODE_VERSION}-slim AS deps
ARG B_FLAG
ARG VITE_API_URL
# Install
RUN --mount=type=secret,id=NPM_TOKEN,env=NPM_TOKEN --mount=type=cache,target=/root/.npm \
    npm ci

from node:24-slim as runtime
ARG B_FLAG
ARG VITE_API_URL
RUN --mount=type=secret,id=NPM_TOKEN,env=NPM_TOKEN echo RUN inside a continued line \
    RUN not an instruction
CMD ["node", "server.js"]`
	assert.Equal(t, want, got)
	assert.NotContains(t, got, "s3cr3t")
	assert.Empty(t, secretsDeclaredAsArg(got, []string{"NPM_TOKEN"}))
}

// Without variables, the Dockerfile is the generator's; a name ARG does not
// take is left out of it, and is still a build argument.
func TestBuildVariablesLeaveTheDockerfileAlone(t *testing.T) {
	assert.Equal(t, generatedDockerfile, withBuildVariables(generatedDockerfile, nil, nil))

	value := "x"
	got := withBuildVariables("FROM alpine\nRUN true", map[string]*string{"my-var": &value}, nil)
	assert.Equal(t, "FROM alpine\nRUN true", got)
}

// The build's context is the directory a Dockerfile was generated for, when it
// is not the checkout.
func TestABuildIsGivenItsContextDirectory(t *testing.T) {
	data := secretBuildData(t, "FROM scratch\n")
	args, _ := buildxInvocation(data, "", "", nil)
	assert.Equal(t, data.CheckoutDir, args[len(args)-1])

	data.ContextDir = filepath.Join(data.CheckoutDir, "apps", "api")
	args, _ = buildxInvocation(data, "", "", nil)
	assert.Equal(t, data.ContextDir, args[len(args)-1])
}

// Generating a Dockerfile says in the deployment's log what it detected, and
// which variables its steps are given - their names, never a value.
func TestGeneratingADockerfileLogsWhatItDetected(t *testing.T) {
	checkout := t.TempDir()
	app := filepath.Join(checkout, "apps", "web")
	assert.NoError(t, os.MkdirAll(app, 0o755))
	assert.NoError(t, os.WriteFile(filepath.Join(app, "package.json"),
		[]byte(`{"scripts": {"build": "vite build", "start": "node server.js"}}`), 0o600))
	assert.NoError(t, os.WriteFile(filepath.Join(app, "package-lock.json"), []byte(`{}`), 0o600))
	apiURL := "https://api.example.com"
	logStore := tasklog.NewLocalStore("test")
	data := &imageBuildData{
		ImageBuildReq: &imagebuildservice.ImageBuildReq{
			TaskExecData: &queue.TaskExecData{LogStore: logStore},
			CheckoutDir:  checkout,
			Dockerfile: entity.DeploymentDockerfile{
				Source: base.DockerfileSourceAuto, Path: "Dockerfile", ScanPath: "apps/web",
			},
		},
		EnvVars:       map[string]*string{"VITE_API_URL": &apiURL},
		SecretEnvVars: map[string]string{"NPM_TOKEN": "s3cr3t"},
	}

	assert.NoError(t, (&service{}).prepareDockerfile(context.Background(), data))

	frames, err := logStore.GetLocalData(context.Background(), 0)
	assert.NoError(t, err)
	var text strings.Builder
	for _, frame := range frames {
		if frame.Type != tasklog.LogTypeDebug {
			text.WriteString(frame.Data + "\n")
		}
	}
	assert.Contains(t, text.String(), "Detected Node project")
	assert.Contains(t, text.String(), "Build command   : npm run build")
	assert.Contains(t, text.String(), "VITE_API_URL; as secrets: NPM_TOKEN")
	assert.NotContains(t, text.String(), apiURL)
	assert.NotContains(t, text.String(), "s3cr3t")

	written, err := os.ReadFile(filepath.Join(checkout, "Dockerfile"))
	assert.NoError(t, err)
	assert.Contains(t, string(written), "ARG VITE_API_URL\n")
	assert.Contains(t, string(written), "--mount=type=secret,id=NPM_TOKEN,env=NPM_TOKEN")
	assert.Equal(t, filepath.Join(checkout, "apps", "web"), data.ContextDir)
}

// End to end, with docker's own builder: a Node.js app in a subdirectory, its
// Dockerfile generated, built with the subdirectory as its context, and its
// build step given a plain variable and a secret. It builds an image, so it
// runs only when HIVEPAAS_DOCKER_TESTS=1.
func TestAGeneratedDockerfileGivesTheBuildItsVariables(t *testing.T) {
	if os.Getenv("HIVEPAAS_DOCKER_TESTS") != "1" {
		t.Skip("set HIVEPAAS_DOCKER_TESTS=1 to build an image with docker")
	}
	checkout := t.TempDir()
	app := filepath.Join(checkout, "apps", "web")
	assert.NoError(t, os.MkdirAll(app, 0o755))
	// The build writes what it was given; the secret's length, not its value.
	build := `node -e "require('fs').writeFileSync('built.txt', ` +
		`process.env.VITE_API_URL + ' ' + (process.env.NPM_TOKEN || '').length)"`
	assert.NoError(t, os.WriteFile(filepath.Join(app, "package.json"), []byte(
		`{"name": "web", "private": true, "scripts": {"build": `+strconv.Quote(build)+`, "start": "node -v"}}`), 0o600))
	assert.NoError(t, os.WriteFile(filepath.Join(app, "package-lock.json"), []byte(
		`{"name": "web", "lockfileVersion": 3, "requires": true, "packages": {"": {"name": "web"}}}`), 0o600))
	// The repository's root has a package.json of its own, which the build
	// must not be given
	assert.NoError(t, os.WriteFile(filepath.Join(checkout, "package.json"), []byte(`{"name": "root"}`), 0o600))

	apiURL := "https://api.example.com"
	data := &imageBuildData{
		ImageBuildReq: &imagebuildservice.ImageBuildReq{
			TaskExecData: &queue.TaskExecData{LogStore: tasklog.NewNullStore()},
			CheckoutDir:  checkout,
			Dockerfile: entity.DeploymentDockerfile{
				Source: base.DockerfileSourceAuto, Path: "Dockerfile", ScanPath: "apps/web",
			},
		},
		ImageTags:     []string{"hp-build-vars-test:latest"},
		EnvVars:       map[string]*string{"VITE_API_URL": &apiURL},
		SecretEnvVars: map[string]string{"NPM_TOKEN": "s3cr3t-hp-value"},
	}
	assert.NoError(t, (&service{}).prepareDockerfile(context.Background(), data))
	assert.NoError(t, checkSecretsNotDeclaredAsArg(data))
	t.Cleanup(func() { _ = exec.Command("docker", "rmi", "-f", data.ImageTags[0]).Run() })

	args, env := buildxInvocation(data, "", "", os.Environ())
	cmd := exec.Command("docker", args...)
	cmd.Dir, cmd.Env = data.CheckoutDir, env
	out, err := cmd.CombinedOutput()
	if !assert.NoError(t, err, string(out)) {
		return
	}

	built, err := exec.Command("docker", "run", "--rm", data.ImageTags[0], "cat", "built.txt").CombinedOutput()
	assert.NoError(t, err)
	assert.Equal(t, "https://api.example.com 15", string(built))
	history, err := exec.Command("docker", "history", "--no-trunc", data.ImageTags[0]).CombinedOutput()
	assert.NoError(t, err)
	assert.NotContains(t, string(history), "s3cr3t")
}

// A Dockerfile is written and read in the checkout only: a path out of it, or a
// link the repository holds - named as the Dockerfile or on the way to it -
// is refused, and the file it names on the node is left as it was.
func TestADockerfileIsNeverWrittenOutOfTheCheckout(t *testing.T) {
	cases := map[string]struct {
		source base.DockerfileSource
		path   string
		// What the repository holds, a link to the node's file or directory
		link, linkTo string
	}{
		"a path out of it, with content":     {base.DockerfileSourceManual, "../victim/Dockerfile", "", ""},
		"a link as the Dockerfile, content":  {base.DockerfileSourceManual, "Dockerfile", "Dockerfile", "file"},
		"a link as the Dockerfile, auto":     {base.DockerfileSourceAuto, "Dockerfile", "Dockerfile", "file"},
		"a link on the way to it, content":   {base.DockerfileSourceManual, "docker/Dockerfile", "docker", "dir"},
		"a link on the way to it, auto":      {base.DockerfileSourceAuto, "docker/Dockerfile", "docker", "dir"},
		"a reserved directory that is links": {base.DockerfileSourceManual, ".hivepaas/Dockerfile", ".hivepaas", "dir"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			parent := t.TempDir()
			checkout := filepath.Join(parent, "checkout")
			victimDir := filepath.Join(parent, "victim")
			victim := filepath.Join(victimDir, "Dockerfile")
			assert.NoError(t, os.MkdirAll(checkout, 0o755))
			assert.NoError(t, os.MkdirAll(victimDir, 0o755))
			assert.NoError(t, os.WriteFile(victim, []byte("the node's"), 0o600))
			assert.NoError(t, os.WriteFile(filepath.Join(checkout, "index.html"), []byte("<p>x</p>"), 0o600))
			switch tc.linkTo {
			case "file":
				assert.NoError(t, os.Symlink(victim, filepath.Join(checkout, tc.link)))
			case "dir":
				assert.NoError(t, os.Symlink(victimDir, filepath.Join(checkout, tc.link)))
			}
			data := &imageBuildData{ImageBuildReq: &imagebuildservice.ImageBuildReq{
				TaskExecData: &queue.TaskExecData{LogStore: tasklog.NewNullStore()},
				CheckoutDir:  checkout,
				Dockerfile:   entity.DeploymentDockerfile{Source: tc.source, Path: tc.path, Content: "FROM scratch"},
			}}

			err := (&service{}).prepareDockerfile(context.Background(), data)

			assert.ErrorIs(t, err, hperrors.ErrArgumentInvalid)
			content, _ := os.ReadFile(victim)
			assert.Equal(t, "the node's", string(content))
		})
	}
}

// The repository's own Dockerfile is read where it is, a link within the
// checkout followed; one that leads out of it is refused, not found.
func TestARepositorysDockerfileIsReadInTheCheckoutOnly(t *testing.T) {
	parent := t.TempDir()
	checkout := filepath.Join(parent, "checkout")
	assert.NoError(t, os.MkdirAll(filepath.Join(checkout, "docker"), 0o755))
	assert.NoError(t, os.WriteFile(filepath.Join(checkout, "docker", "Dockerfile.prod"), []byte("FROM scratch"), 0o600))
	assert.NoError(t, os.Symlink("docker/Dockerfile.prod", filepath.Join(checkout, "Dockerfile")))
	assert.NoError(t, os.WriteFile(filepath.Join(parent, "secret"), []byte("FROM scratch"), 0o600))
	assert.NoError(t, os.Symlink(filepath.Join(parent, "secret"), filepath.Join(checkout, "Outside")))
	prepare := func(path string) error {
		return (&service{}).prepareDockerfile(context.Background(), &imageBuildData{
			ImageBuildReq: &imagebuildservice.ImageBuildReq{
				TaskExecData: &queue.TaskExecData{LogStore: tasklog.NewNullStore()},
				CheckoutDir:  checkout,
				Dockerfile:   entity.DeploymentDockerfile{Source: base.DockerfileSourceManual, Path: path},
			},
		})
	}

	assert.NoError(t, prepare("Dockerfile"))
	assert.NoError(t, prepare("docker/Dockerfile.prod"))
	assert.ErrorIs(t, prepare("Outside"), hperrors.ErrArgumentInvalid)
	assert.ErrorIs(t, prepare("../secret"), hperrors.ErrArgumentInvalid)
	assert.ErrorIs(t, prepare("Missing"), hperrors.ErrNotFound)
}
