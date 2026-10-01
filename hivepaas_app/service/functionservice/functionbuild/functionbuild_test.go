package functionbuild

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
)

var images = map[string]string{
	"node24":      "ghcr.io/hivepaas/function-runtime-node24:1.0.0@sha256:node",
	"python313":   "ghcr.io/hivepaas/function-runtime-python313:1.0.0@sha256:python",
	"go127":       "ghcr.io/hivepaas/function-runtime-go127:1.0.0@sha256:go",
	"go127-build": "ghcr.io/hivepaas/function-runtime-go127-build:1.0.0@sha256:gobuild",
}

func source(runtime base.FunctionRuntime) *entity.DeploymentFunctionSource {
	file, handler := runtime.DefaultEntrypoint()
	return &entity.DeploymentFunctionSource{
		Runtime:        runtime,
		Contract:       base.FunctionContractV1,
		Entrypoint:     entity.FunctionEntrypoint{File: file, Handler: handler},
		Timeout:        timeutil.Duration(30 * time.Second),
		MaxConcurrency: 16,
		MaxBodySize:    6 * unit.MB,
	}
}

// sourceDir is a function's source holding files, by path.
func sourceDir(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		path := filepath.Join(dir, name)
		assert.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		assert.NoError(t, os.WriteFile(path, []byte(content), 0o600))
	}
	return dir
}

// A Node.js function's libraries are installed from its manifest, its lock file
// and its .npmrc alone, so that a change to its code alone does not install
// them again; the build's secrets reach the install, its other variables too.
func TestANodeFunctionInstallsItsLibrariesBeforeItsCode(t *testing.T) {
	src := source(base.FunctionRuntimeNode24)
	src.SystemPackages = []string{"ffmpeg", "libpq5"}
	dir := sourceDir(t, map[string]string{
		"index.js": "export default () => {}", "package.json": `{"dependencies": {"pg": "^8"}}`,
		"package-lock.json": "{}", ".npmrc": "//registry.npmjs.org/:_authToken=${NPM_TOKEN}",
	})

	resp, err := Dockerfile(&DockerfileReq{
		Source: src, Images: images, SourceDir: dir,
		BuildArgs: []string{"NPM_CONFIG_LOGLEVEL"}, BuildSecrets: []string{"NPM_TOKEN", "GH_TOKEN"},
	})

	assert.NoError(t, err)
	assert.Equal(t, `# The Dockerfile HivePaaS writes for a function: runtime node24, contract v1.
FROM ghcr.io/hivepaas/function-runtime-node24:1.0.0@sha256:node AS base
USER root
RUN apt-get update \
    && apt-get install -y --no-install-recommends ffmpeg libpq5 \
    && rm -rf /var/lib/apt/lists/*
USER hivepaas

FROM base AS deps
COPY --chown=hivepaas:hivepaas package.json package-lock.json .npmrc /app/
ARG NPM_CONFIG_LOGLEVEL
RUN --mount=type=secret,id=GH_TOKEN,env=GH_TOKEN --mount=type=secret,id=NPM_TOKEN,env=NPM_TOKEN \
    hivepaas-runtime deps

FROM deps
COPY --chown=hivepaas:hivepaas . /app/
ENV HP_FN_ENTRYPOINT="index.js" \
    HP_FN_HANDLER="default" \
    HP_FN_TIMEOUT_MS=30000 \
    HP_FN_MAX_CONCURRENCY=16 \
    HP_FN_MAX_BODY_SIZE=6291456
`, resp.Content)
	assert.Empty(t, resp.Notes)
}

// Without a manifest there is nothing to install, and no package is no root.
func TestAFunctionWithoutLibrariesHasNoInstallStep(t *testing.T) {
	dir := sourceDir(t, map[string]string{"main.py": "def handler(req, ctx): pass"})

	resp, err := Dockerfile(&DockerfileReq{Source: source(base.FunctionRuntimePython313), Images: images,
		SourceDir: dir})

	assert.NoError(t, err)
	assert.Equal(t, `# The Dockerfile HivePaaS writes for a function: runtime python313, contract v1.
FROM ghcr.io/hivepaas/function-runtime-python313:1.0.0@sha256:python AS base

FROM base
COPY --chown=hivepaas:hivepaas . /app/
ENV HP_FN_ENTRYPOINT="main.py" \
    HP_FN_HANDLER="handler" \
    HP_FN_TIMEOUT_MS=30000 \
    HP_FN_MAX_CONCURRENCY=16 \
    HP_FN_MAX_BODY_SIZE=6291456
`, resp.Content)
	assert.Empty(t, resp.Notes)
}

// A manifest without its lock file is installed all the same, and the build log
// says the versions are being resolved.
func TestALockFileThatIsMissingIsNoted(t *testing.T) {
	dir := sourceDir(t, map[string]string{"main.py": "", "requirements.txt": "requests==2.32.3\n"})

	resp, err := Dockerfile(&DockerfileReq{Source: source(base.FunctionRuntimePython313), Images: images,
		SourceDir: dir})

	assert.NoError(t, err)
	assert.Contains(t, resp.Content, "COPY --chown=hivepaas:hivepaas requirements.txt /app/\nRUN hivepaas-runtime deps\n")
	assert.Len(t, resp.Notes, 1)
	assert.Contains(t, resp.Notes[0], "requirements.lock")
}

// A library from the function's own files is not in the install step's copy:
// the code goes in first, whole.
func TestLibrariesFromTheFunctionsOwnFilesAreInstalledWithItsCode(t *testing.T) {
	for name, files := range map[string]map[string]string{
		"an npm file: dependency": {"package.json": `{"dependencies": {"lib": "file:./lib"}}`,
			"package-lock.json": "{}"},
		"a wheel next to requirements.txt": {"requirements.txt": "./wheels/lib-1.0-py3-none-any.whl\n",
			"requirements.lock": ""},
		"a Go module replaced by a directory": {"go.mod": "module fn\n\nreplace example.com/lib => ./lib\n",
			"go.sum": ""},
	} {
		t.Run(name, func(t *testing.T) {
			runtime := base.FunctionRuntimeNode24
			switch {
			case files["requirements.txt"] != "":
				runtime = base.FunctionRuntimePython313
			case files["go.mod"] != "":
				runtime = base.FunctionRuntimeGo127
			}

			resp, err := Dockerfile(&DockerfileReq{Source: source(runtime), Images: images,
				SourceDir: sourceDir(t, files)})

			assert.NoError(t, err)
			assert.Contains(t, resp.Content, "COPY --chown=hivepaas:hivepaas . /app/\nRUN hivepaas-runtime deps\n")
			assert.Equal(t, 1, strings.Count(resp.Content, "COPY --chown=hivepaas:hivepaas . /app/"))
			assert.Len(t, resp.Notes, 1)
		})
	}
}

// A Go function is compiled in the build image and runs on the slim one, which
// gets the Debian packages too.
func TestAGoFunctionIsCompiledThenRunOnTheSlimImage(t *testing.T) {
	src := source(base.FunctionRuntimeGo127)
	src.SystemPackages = []string{"ca-certificates"}
	dir := sourceDir(t, map[string]string{"go.mod": "module fn\n", "go.sum": "", "fn.go": "package fn"})

	resp, err := Dockerfile(&DockerfileReq{Source: src, Images: images, SourceDir: dir,
		BuildSecrets: []string{"GOPROXY"}})

	assert.NoError(t, err)
	assert.Equal(t, `# The Dockerfile HivePaaS writes for a function: runtime go127, contract v1.
FROM ghcr.io/hivepaas/function-runtime-go127-build:1.0.0@sha256:gobuild AS base
USER root
RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*
USER hivepaas

FROM base AS deps
COPY --chown=hivepaas:hivepaas go.mod go.sum /app/
RUN --mount=type=secret,id=GOPROXY,env=GOPROXY \
    hivepaas-runtime deps

FROM deps AS build
COPY --chown=hivepaas:hivepaas . /app/
ENV HP_FN_ENTRYPOINT="." \
    HP_FN_HANDLER="Handle"
RUN --mount=type=secret,id=GOPROXY,env=GOPROXY \
    hivepaas-runtime build

FROM ghcr.io/hivepaas/function-runtime-go127:1.0.0@sha256:go
USER root
RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*
USER hivepaas
COPY --from=build --chown=hivepaas:hivepaas /app /app
ENV HP_FN_ENTRYPOINT="." \
    HP_FN_HANDLER="Handle" \
    HP_FN_TIMEOUT_MS=30000 \
    HP_FN_MAX_CONCURRENCY=16 \
    HP_FN_MAX_BODY_SIZE=6291456
`, resp.Content)
}

// A Node.js export may be named with a '$', which a Dockerfile would expand.
func TestAHandlerNameStaysAsItIs(t *testing.T) {
	src := source(base.FunctionRuntimeNode24)
	src.Entrypoint.Handler = "$handle"

	resp, err := Dockerfile(&DockerfileReq{Source: src, Images: images, SourceDir: t.TempDir()})

	assert.NoError(t, err)
	assert.Contains(t, resp.Content, `HP_FN_HANDLER="\$handle"`)
}

// A runtime the release has no image for cannot be built, nor can a package
// that is not a package's name.
func TestWhatCannotBeBuiltIsRefused(t *testing.T) {
	_, err := Dockerfile(&DockerfileReq{Source: source(base.FunctionRuntimeGo127),
		Images: map[string]string{"go127": images["go127"]}, SourceDir: t.TempDir()})
	assert.ErrorIs(t, err, hperrors.ErrFunctionRuntimeUnavailable)

	src := source(base.FunctionRuntimeNode24)
	src.SystemPackages = []string{"curl; rm -rf /"}
	_, err = Dockerfile(&DockerfileReq{Source: src, Images: images, SourceDir: t.TempDir()})
	assert.ErrorIs(t, err, hperrors.ErrValueInvalid)
}

func TestInlineCodeIsWrittenWhereItSays(t *testing.T) {
	dir := t.TempDir()

	err := WriteInlineCode(dir, &entity.FunctionInlineCode{Files: []*entity.FunctionFile{
		{Path: "index.js", Content: "export default () => {}"},
		{Path: "lib/util/strings.js", Content: "export const a = 1"},
	}})

	assert.NoError(t, err)
	content, _ := os.ReadFile(filepath.Join(dir, "lib", "util", "strings.js"))
	assert.Equal(t, "export const a = 1", string(content))
	info, _ := os.Stat(filepath.Join(dir, "index.js"))
	assert.Equal(t, os.FileMode(0o644), info.Mode().Perm())
}

func TestInlineCodeOutsideTheFunctionIsNotWritten(t *testing.T) {
	for _, path := range []string{"../escape.js", "/etc/escape.js", ".hivepaas/Dockerfile", ""} {
		dir := filepath.Join(t.TempDir(), "fn")
		assert.NoError(t, os.MkdirAll(dir, 0o755))

		err := WriteInlineCode(dir, &entity.FunctionInlineCode{Files: []*entity.FunctionFile{
			{Path: path, Content: "x"},
		}})

		assert.ErrorIs(t, err, hperrors.ErrValueInvalid, path)
		_, statErr := os.Stat(filepath.Join(filepath.Dir(dir), "escape.js"))
		assert.True(t, os.IsNotExist(statErr), path)
	}
}

// Inline code has no commit: its image is tagged after its content, the
// Dockerfile included, so that the same function builds to the same tag.
func TestTheSameFunctionHasTheSameContentHash(t *testing.T) {
	code := &entity.FunctionInlineCode{Files: []*entity.FunctionFile{
		{Path: "index.js", Content: "a"}, {Path: "lib.js", Content: "b"},
	}}
	reordered := &entity.FunctionInlineCode{Files: []*entity.FunctionFile{code.Files[1], code.Files[0]}}

	hash := ContentHash(code, "FROM x")

	assert.Len(t, hash, 64)
	assert.Equal(t, hash, ContentHash(reordered, "FROM x"))
	assert.NotEqual(t, hash, ContentHash(code, "FROM y"))
	assert.NotEqual(t, hash, ContentHash(&entity.FunctionInlineCode{Files: []*entity.FunctionFile{
		{Path: "index.js", Content: "a"}, {Path: "lib.js", Content: "c"},
	}}, "FROM x"))
	assert.NotEqual(t, ContentHash(&entity.FunctionInlineCode{Files: []*entity.FunctionFile{
		{Path: "ab", Content: "c"},
	}}, ""), ContentHash(&entity.FunctionInlineCode{Files: []*entity.FunctionFile{
		{Path: "a", Content: "bc"},
	}}, ""))
}
