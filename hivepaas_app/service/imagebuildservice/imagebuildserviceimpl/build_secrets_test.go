package imagebuildserviceimpl

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagebuildservice"
)

func secretBuildData(t *testing.T, dockerfile string) *imageBuildData {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "Dockerfile"), []byte(dockerfile), 0o600); err != nil {
		t.Fatal(err)
	}
	nodeEnv := "production"
	return &imageBuildData{
		ImageBuildReq: &imagebuildservice.ImageBuildReq{
			CheckoutDir: dir,
			Dockerfile:  entity.DeploymentDockerfile{Path: "Dockerfile"},
		},
		ImageTags:     []string{"hp-build-secrets-test:latest"},
		EnvVars:       map[string]*string{"NODE_ENV": &nodeEnv},
		SecretEnvVars: map[string]string{"NPM_TOKEN": "s3cr3t-hp-value", "PIP_TOKEN": "other-s3cr3t"},
	}
}

// A secret build variable reaches buildx as a BuildKit secret whose value is in
// the process environment: never on the command line, which anyone listing the
// node's processes can read, and never as a build argument, which the image keeps.
func TestASecretBuildVariableNeverReachesTheCommandLine(t *testing.T) {
	data := secretBuildData(t, "FROM scratch\n")

	args, env := buildxInvocation(data, "hivepaas_builder", "", []string{"PATH=/usr/bin"})

	joined := strings.Join(args, " ")
	assert.NotContains(t, joined, "s3cr3t")
	assert.Contains(t, joined, "--build-arg NODE_ENV=production")
	assert.Contains(t, joined, "--secret id=NPM_TOKEN,env=HP_BUILD_SECRET_0")
	assert.Contains(t, joined, "--secret id=PIP_TOKEN,env=HP_BUILD_SECRET_1")
	assert.Contains(t, env, "HP_BUILD_SECRET_0=s3cr3t-hp-value")
	assert.Contains(t, env, "HP_BUILD_SECRET_1=other-s3cr3t")
	assert.Contains(t, env, "PATH=/usr/bin")
}

// Which names a Dockerfile declares with ARG, as Docker reads the instruction.
func TestSecretsDeclaredAsArg(t *testing.T) {
	secrets := []string{"NPM_TOKEN", "PIP_TOKEN"}
	cases := map[string]struct {
		dockerfile string
		want       []string
	}{
		"declared":                   {"FROM node\nARG NPM_TOKEN\nRUN npm ci\n", []string{"NPM_TOKEN"}},
		"with a default":             {"ARG NPM_TOKEN=none\n", []string{"NPM_TOKEN"}},
		"the keyword in lower case":  {"arg NPM_TOKEN\n", []string{"NPM_TOKEN"}},
		"several on one line":        {"ARG A B=1 PIP_TOKEN\n", []string{"PIP_TOKEN"}},
		"on a continued line":        {"ARG A \\\n    NPM_TOKEN\n", []string{"NPM_TOKEN"}},
		"both, in order":             {"ARG PIP_TOKEN\nARG NPM_TOKEN\n", []string{"NPM_TOKEN", "PIP_TOKEN"}},
		"in a comment":               {"# ARG NPM_TOKEN\n", nil},
		"inside another instruction": {"RUN echo ARG NPM_TOKEN\n", nil},
		"another case of the name":   {"ARG npm_token\n", nil},
		"only plain variables":       {"ARG NODE_ENV\n", nil},
		"read through a secret mount": {
			"RUN --mount=type=secret,id=NPM_TOKEN,env=NPM_TOKEN npm ci\n", nil,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, secretsDeclaredAsArg(tc.dockerfile, secrets))
		})
	}
}

// A Dockerfile that declares a secret variable with ARG would write its value
// into the image: the build stops before it starts, saying which one and how to
// read it instead.
func TestABuildRefusesADockerfileDeclaringASecretAsArg(t *testing.T) {
	data := secretBuildData(t, "FROM node:22\nARG NPM_TOKEN\nRUN npm ci\n")

	err := checkSecretsNotDeclaredAsArg(data)

	// Which variables, and how to read them instead, are in the error's message.
	assert.ErrorIs(t, err, hperrors.ErrBuildSecretDeclaredAsArg)
}

// The same build reading its secret through a secret mount goes ahead.
func TestABuildReadingItsSecretsThroughAMountGoesAhead(t *testing.T) {
	data := secretBuildData(t, "FROM node:22\nARG NODE_ENV\n"+
		"RUN --mount=type=secret,id=NPM_TOKEN,env=NPM_TOKEN npm ci\n")

	assert.NoError(t, checkSecretsNotDeclaredAsArg(data))
}

// End to end, with docker's own builder: the secret reaches the RUN that mounts
// it, and the image keeps no trace of it. It builds an image, so it runs only
// when HIVEPAAS_DOCKER_TESTS=1.
func TestASecretMountedInABuildLeavesNoTraceInTheImage(t *testing.T) {
	if os.Getenv("HIVEPAAS_DOCKER_TESTS") != "1" {
		t.Skip("set HIVEPAAS_DOCKER_TESTS=1 to build an image with docker")
	}
	data := secretBuildData(t, "FROM alpine:3.20\n"+
		"RUN --mount=type=secret,id=NPM_TOKEN,env=NPM_TOKEN "+
		// The value's length, not the value: the RUN line itself is in the history.
		"sh -c 'test ${#NPM_TOKEN} = 15 && echo seen > /seen'\n")
	t.Cleanup(func() { _ = exec.Command("docker", "rmi", "-f", data.ImageTags[0]).Run() })

	// No --builder: docker's default builder, not one this test would create.
	args, env := buildxInvocation(data, "", "", os.Environ())
	cmd := exec.Command("docker", args...)
	cmd.Dir, cmd.Env = data.CheckoutDir, env
	out, err := cmd.CombinedOutput()
	if !assert.NoError(t, err, string(out)) {
		return
	}

	history, err := exec.Command("docker", "history", "--no-trunc", data.ImageTags[0]).CombinedOutput()
	assert.NoError(t, err)
	assert.NotContains(t, string(history), "s3cr3t")
	inspect, err := exec.Command("docker", "image", "inspect", data.ImageTags[0]).CombinedOutput()
	assert.NoError(t, err)
	assert.NotContains(t, string(inspect), "s3cr3t")
	seen, err := exec.Command("docker", "run", "--rm", data.ImageTags[0], "cat", "/seen").CombinedOutput()
	assert.NoError(t, err)
	assert.Equal(t, "seen\n", string(seen))
}
