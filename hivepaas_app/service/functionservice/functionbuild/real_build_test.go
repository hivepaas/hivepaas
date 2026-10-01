package functionbuild

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

// realFunctions are a function of each runtime that uses a library from its
// package registry and a Debian package, so that a build installs both.
var realFunctions = map[base.FunctionRuntime][]*entity.FunctionFile{
	base.FunctionRuntimeNode24: {
		{Path: "package.json", Content: `{"type": "module", "dependencies": {"ms": "2.1.3"}}`},
		{Path: "index.js", Content: `import { execSync } from 'node:child_process'
import ms from 'ms'

export default () => ({ body: { library: String(ms('1m')), tool: execSync('jq --version').toString().trim() } })
`},
	},
	base.FunctionRuntimePython313: {
		{Path: "requirements.txt", Content: "six==1.17.0\n"},
		{Path: "main.py", Content: `import subprocess

import six


def handler(req, ctx):
    tool = subprocess.run(["jq", "--version"], capture_output=True, text=True).stdout.strip()
    return {"body": {"library": six.__version__, "tool": tool}}
`},
	},
	base.FunctionRuntimeGo127: {
		{Path: "go.mod", Content: "module example.com/fn\n\ngo 1.27\n"},
		{Path: "fn.go", Content: `package fn

import (
	"context"
	"os/exec"
	"strings"

	"github.com/google/uuid"
	"github.com/hivepaas/function-runtimes/hivepaas"
)

func Handle(ctx context.Context, req *hivepaas.Request) (*hivepaas.Response, error) {
	tool, err := exec.Command("jq", "--version").Output()
	if err != nil {
		return nil, err
	}
	return hivepaas.JSON(200, map[string]string{
		"library": uuid.NewSHA1(uuid.NameSpaceURL, []byte("hivepaas")).String()[:8],
		"tool":    strings.TrimSpace(string(tool)),
	})
}
`},
	},
}

var realLibraryValues = map[base.FunctionRuntime]string{
	base.FunctionRuntimeNode24:    "60000",
	base.FunctionRuntimePython313: "1.17.0",
	base.FunctionRuntimeGo127:     uuidOfHivepaas,
}

// uuidOfHivepaas is the first part of uuid.NewSHA1(uuid.NameSpaceURL, "hivepaas").
const uuidOfHivepaas = "a7b554de"

// End to end, with docker's own builder and the release's runtime images: a
// function of each runtime is built from the Dockerfile written for it, with
// its library, a Debian package and a build secret, then called once. It pulls
// the images and the libraries, so it runs only when HIVEPAAS_DOCKER_TESTS=1.
func TestAFunctionOfEachRuntimeBuildsOnTheReleasesImages(t *testing.T) {
	if os.Getenv("HIVEPAAS_DOCKER_TESTS") != "1" {
		t.Skip("set HIVEPAAS_DOCKER_TESTS=1 to build functions with docker")
	}
	for runtime, files := range realFunctions {
		t.Run(string(runtime), func(t *testing.T) {
			dir := t.TempDir()
			assert.NoError(t, WriteInlineCode(dir, &entity.FunctionInlineCode{Files: files}))
			src := source(runtime)
			src.SystemPackages = []string{"jq"}
			resp, err := Dockerfile(&DockerfileReq{Source: src, Images: base.BetaVersion.FunctionRuntimes,
				SourceDir: dir, BuildArgs: []string{"HP_TEST_ARG"}, BuildSecrets: []string{"HP_TEST_SECRET"}})
			if !assert.NoError(t, err) {
				return
			}
			assert.NoError(t, os.MkdirAll(filepath.Join(dir, base.FunctionReservedDir), 0o755))
			assert.NoError(t, os.WriteFile(filepath.Join(dir, DockerfilePath), []byte(resp.Content), 0o600))
			tag := "hivepaas-test-function-" + string(runtime)
			t.Cleanup(func() { _ = exec.Command("docker", "rmi", "-f", tag).Run() })

			// No --builder: docker's default builder, not one this test would create.
			build := exec.Command("docker", "buildx", "build", "--load", "-f", DockerfilePath, "-t", tag,
				"--build-arg", "HP_TEST_ARG=arg", "--secret", "id=HP_TEST_SECRET,env=HP_BUILD_SECRET_0", ".")
			build.Dir, build.Env = dir, append(os.Environ(), "HP_BUILD_SECRET_0=s3cr3t-value")
			out, err := build.CombinedOutput()
			if !assert.NoError(t, err, string(out)) {
				return
			}

			call := exec.Command("docker", "run", "--rm", "-i", "--network", "none", tag, "hivepaas-runtime", "invoke")
			call.Stdin = strings.NewReader(`{}`)
			out, err = call.Output()
			if !assert.NoError(t, err, string(out)) {
				return
			}
			lines := strings.Split(strings.TrimSpace(string(out)), "\n")
			var result struct {
				Status int    `json:"status"`
				Body   string `json:"body"`
			}
			assert.NoError(t, json.Unmarshal([]byte(strings.TrimPrefix(lines[len(lines)-1], "#hivepaas-result ")),
				&result))
			body, _ := base64.StdEncoding.DecodeString(result.Body)
			var got map[string]string
			assert.NoError(t, json.Unmarshal(body, &got), string(body))
			assert.Equal(t, 200, result.Status)
			assert.Equal(t, realLibraryValues[runtime], got["library"])
			assert.True(t, strings.HasPrefix(got["tool"], "jq-"), got["tool"])

			history, err := exec.Command("docker", "history", "--no-trunc", tag).CombinedOutput()
			assert.NoError(t, err)
			assert.NotContains(t, string(history), "s3cr3t-value")
		})
	}
}
