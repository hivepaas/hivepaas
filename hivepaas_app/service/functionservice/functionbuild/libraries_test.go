package functionbuild

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func librariesOf(t *testing.T, runtime base.FunctionRuntime, packages []string, files map[string]string,
) *LibrariesResp {
	t.Helper()
	src := source(runtime)
	src.SystemPackages = packages
	resp, err := Libraries(&DockerfileReq{Source: src, Images: images, SourceDir: sourceDir(t, files),
		BuildArgs: []string{"NPM_CONFIG_LOGLEVEL"}, BuildSecrets: []string{"NPM_TOKEN"}})
	assert.NoError(t, err)
	return resp
}

// A test run's libraries are the function's Dockerfile up to its install, with
// no code and no limits: the code goes into the test run's container, and the
// limits into its environment.
func TestALibrariesImageIsTheInstallAlone(t *testing.T) {
	resp := librariesOf(t, base.FunctionRuntimeNode24, []string{"ffmpeg"}, map[string]string{
		"index.js": "export default () => {}", "package.json": `{"dependencies": {"ms": "2.1.3"}}`,
		"package-lock.json": "{}",
	})

	assert.Equal(t, `# The libraries of a function, for its test runs: runtime node24, contract v1.
FROM ghcr.io/hivepaas/function-runtime-node24:1.0.0@sha256:node AS base
USER root
RUN apt-get update \
    && apt-get install -y --no-install-recommends ffmpeg \
    && rm -rf /var/lib/apt/lists/*
USER hivepaas

FROM base AS deps
COPY --chown=hivepaas:hivepaas package.json package-lock.json /app/
ARG NPM_CONFIG_LOGLEVEL
RUN --mount=type=secret,id=NPM_TOKEN,env=NPM_TOKEN \
    hivepaas-runtime deps
`, resp.Dockerfile)
	assert.True(t, strings.HasPrefix(resp.Image, LibrariesRepo+":"), resp.Image)
	assert.Len(t, strings.TrimPrefix(resp.Image, LibrariesRepo+":"), 24)
}

// A Go function's test run compiles and calls in the build image, which is
// where its libraries are.
func TestAGoFunctionsLibrariesAreInTheBuildImage(t *testing.T) {
	resp := librariesOf(t, base.FunctionRuntimeGo127, nil, map[string]string{
		"fn.go": "package fn", "go.mod": "module example.com/fn\n",
	})

	assert.Contains(t, resp.Dockerfile, "FROM "+images["go127-build"]+" AS base\n")
	assert.Contains(t, resp.Dockerfile, "COPY --chown=hivepaas:hivepaas go.mod /app/\n")
	assert.NotContains(t, resp.Dockerfile, "hivepaas-runtime build")
}

// The image is named after what it is built from: the libraries are installed
// again only when that changes, not when the code does.
func TestALibrariesImageIsNamedAfterWhatItIsBuiltFrom(t *testing.T) {
	files := map[string]string{"index.js": "export default () => 1", "package.json": `{"dependencies": {}}`}
	first := librariesOf(t, base.FunctionRuntimeNode24, nil, files)

	files["index.js"] = "export default () => 2"
	assert.Equal(t, first.Image, librariesOf(t, base.FunctionRuntimeNode24, nil, files).Image)

	files["package.json"] = `{"dependencies": {"ms": "2.1.3"}}`
	manifestChanged := librariesOf(t, base.FunctionRuntimeNode24, nil, files)
	assert.NotEqual(t, first.Image, manifestChanged.Image)

	files["package-lock.json"] = "{}"
	lockAdded := librariesOf(t, base.FunctionRuntimeNode24, nil, files)
	assert.NotEqual(t, manifestChanged.Image, lockAdded.Image)

	assert.NotEqual(t, lockAdded.Image, librariesOf(t, base.FunctionRuntimeNode24, []string{"jq"}, files).Image)
}

// A manifest that names the function's own files installs with the code, so
// the code names the image too.
func TestLibrariesFromTheFunctionsOwnFilesAreNamedAfterTheCode(t *testing.T) {
	files := map[string]string{"index.js": "1", "lib/a.js": "1",
		"package.json": `{"dependencies": {"lib": "file:./lib"}}`}
	first := librariesOf(t, base.FunctionRuntimeNode24, nil, files)
	assert.Contains(t, first.Dockerfile, "COPY --chown=hivepaas:hivepaas . /app/\n")

	files["lib/a.js"] = "2"
	assert.NotEqual(t, first.Image, librariesOf(t, base.FunctionRuntimeNode24, nil, files).Image)
}

// Without libraries or packages there is nothing to build: the runtime image is
// the test run's.
func TestAFunctionWithoutLibrariesRunsOnItsRuntimesImage(t *testing.T) {
	resp := librariesOf(t, base.FunctionRuntimePython313, nil, map[string]string{"main.py": "def handler(r, c): pass"})

	assert.Empty(t, resp.Dockerfile)
	assert.Equal(t, images["python313"], resp.Image)
}

// The lock file a test run returns is the runtime's: what its install makes,
// and for Go what its build tidies.
func TestEachRuntimeNamesItsLockFiles(t *testing.T) {
	assert.Equal(t, []string{"package-lock.json"}, LockFiles(base.FunctionRuntimeNode24))
	assert.Equal(t, []string{"requirements.lock"}, LockFiles(base.FunctionRuntimePython313))
	assert.Equal(t, []string{"go.mod", "go.sum"}, LockFiles(base.FunctionRuntimeGo127))
}

func TestALibrariesImageNeedsItsRuntimesImage(t *testing.T) {
	src := source(base.FunctionRuntimeNode24)
	_, err := Libraries(&DockerfileReq{Source: src, Images: map[string]string{}, SourceDir: t.TempDir()})

	assert.ErrorIs(t, err, hperrors.ErrFunctionRuntimeUnavailable)
}
