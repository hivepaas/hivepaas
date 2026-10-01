package imagebuildserviceimpl

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"maps"
	"os/exec"
	"slices"
	"sync"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/safego"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
)

func (s *service) buildImageWithDocker(
	ctx context.Context,
	_ database.IDB,
	data *imageBuildData,
) (err error) {
	buildSetting := data.ImageBuildSettings

	_ = data.LogStore.Add(ctx, tasklog.NewOutFrame("Start building image with Docker BuildKit...",
		tasklog.TsNow))

	builderName := base.HivepaasGlobalBuilder
	var res *entity.ImageBuildResourceSettings
	if buildSetting != nil {
		res = &buildSetting.Resources
	}

	if err := s.ensureCustomBuilder(ctx, builderName, res, data.LogStore); err != nil {
		return hperrors.Wrap(err)
	}

	dockerConfigDir, cleanup, err := s.prepareDockerConfigDir(data)
	if err != nil {
		return hperrors.Wrap(err)
	}
	defer cleanup()

	args, env := buildxInvocation(data, builderName, dockerConfigDir, s.calcSafeEnvVars())
	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Dir = data.CheckoutDir
	cmd.Env = env

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return hperrors.Wrap(err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return hperrors.Wrap(err)
	}

	if err := cmd.Start(); err != nil {
		return hperrors.Wrap(err)
	}

	var wg sync.WaitGroup
	wg.Go(func() {
		defer safego.Recover("imagebuild.streamStdout")
		s.streamLogOutput(ctx, data.LogStore, stdout)
	})
	wg.Go(func() {
		defer safego.Recover("imagebuild.streamStderr")
		s.streamLogOutput(ctx, data.LogStore, stderr)
	})
	wg.Wait()

	if err := cmd.Wait(); err != nil {
		return hperrors.Wrap(err)
	}

	return nil
}

func (s *service) streamLogOutput(
	ctx context.Context,
	logStore *tasklog.Store,
	r io.Reader,
) {
	if logStore == nil || r == nil {
		return
	}
	scanner := bufio.NewScanner(r)
	const maxLineLength = 1024 * 1024
	const bufferSize = 32 * 1024
	buf := make([]byte, bufferSize)
	scanner.Buffer(buf, maxLineLength)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		_ = logStore.AddRedacted(ctx, tasklog.NewDebugFrame(line, tasklog.TsNow))
	}
}

// buildSecretEnvPrefix names the environment variables that carry the build's
// secrets to buildx: HP_BUILD_SECRET_0, _1, ... A name of its own, so that a
// secret called PATH does not replace the process's PATH.
const buildSecretEnvPrefix = "HP_BUILD_SECRET_" //nolint:gosec // a name prefix, not a credential

// buildxInvocation is the docker command line of a build, and its environment.
//
// A variable that uses no secret is a build argument. One that uses a secret is
// a BuildKit secret (`--secret id=NAME,env=...`), read in the Dockerfile with
// `RUN --mount=type=secret,id=NAME,env=NAME`: its value is only in the
// environment of the buildx process, where the command line - which anyone
// listing the node's processes can read - does not show it, and the image never
// records it. An empty builderName leaves docker's default builder.
func buildxInvocation(
	data *imageBuildData,
	builderName string,
	dockerConfigDir string,
	safeEnv []string,
) (args, env []string) {
	buildSetting := data.ImageBuildSettings

	args = []string{cmdBuildx, "build"}
	if builderName != "" {
		args = append(args, "--builder", builderName)
	}
	args = append(args, "--load", "--progress=plain", "-f", data.Dockerfile.Path)

	for _, tag := range data.ImageTags {
		args = append(args, "-t", tag)
	}
	for _, name := range slices.Sorted(maps.Keys(data.EnvVars)) {
		if value := data.EnvVars[name]; value != nil {
			args = append(args, "--build-arg", fmt.Sprintf("%s=%s", name, *value))
		}
	}

	env = append(slices.Clone(safeEnv), "DOCKER_BUILDKIT=1")
	if dockerConfigDir != "" {
		env = append(env, "DOCKER_CONFIG="+dockerConfigDir)
	}
	for i, name := range slices.Sorted(maps.Keys(data.SecretEnvVars)) {
		envName := fmt.Sprintf("%s%d", buildSecretEnvPrefix, i)
		args = append(args, "--secret", fmt.Sprintf("id=%s,env=%s", name, envName))
		env = append(env, envName+"="+data.SecretEnvVars[name])
	}

	if data.NoCache || (buildSetting != nil && buildSetting.NoCache) {
		args = append(args, "--no-cache")
	}
	if buildSetting != nil {
		if buildSetting.NoVerbose {
			args = append(args, "--quiet")
		}
		if buildSetting.Resources.ShmSize > 0 {
			args = append(args, "--shm-size", fmt.Sprintf("%d", buildSetting.Resources.ShmSize.Bytes()))
		}
	}

	// CheckoutDir is the build context
	return append(args, data.CheckoutDir), env
}
