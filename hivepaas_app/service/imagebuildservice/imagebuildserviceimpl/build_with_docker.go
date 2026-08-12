package imagebuildserviceimpl

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"sync"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

func (s *service) buildImageWithDocker(
	ctx context.Context,
	_ database.IDB,
	data *imageBuildData,
) (err error) {
	buildSetting := data.ImageBuildSettings

	s.addStepStartLog(ctx, data, "Start building image with Docker BuildKit...")
	defer s.addStepEndLog(ctx, data, timeutil.NowUTC(), err)

	builderName := base.HivepaasGlobalBuilder
	var res *entity.ImageBuildResourceSettings
	if buildSetting != nil {
		res = &buildSetting.Resources
	}

	if err := s.ensureCustomBuilder(ctx, builderName, res); err != nil {
		return apperrors.Wrap(err)
	}

	dockerConfigDir, cleanup, err := s.prepareDockerConfigDir(data)
	if err != nil {
		return apperrors.Wrap(err)
	}
	defer cleanup()

	args := []string{
		"buildx", "build",
		"--builder", builderName,
		"--load",
		"--progress=plain",
		"-f", data.Dockerfile.Path,
	}

	for _, tag := range data.ImageTags {
		args = append(args, "-t", tag)
	}
	for k, v := range data.EnvVars {
		if v != nil {
			args = append(args, "--build-arg", fmt.Sprintf("%s=%s", k, *v))
		}
	}

	if data.NoCache || (buildSetting != nil && buildSetting.NoCache) {
		args = append(args, "--no-cache")
	}

	if buildSetting != nil {
		if buildSetting.NoVerbose {
			args = append(args, "--quiet")
		}
		if res != nil && res.ShmSize > 0 {
			args = append(args, "--shm-size", fmt.Sprintf("%d", res.ShmSize.Bytes()))
		}
	}

	// CheckoutDir is the build context
	args = append(args, data.CheckoutDir)

	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Dir = data.CheckoutDir
	envs := append(s.calcSafeEnvVars(), "DOCKER_BUILDKIT=1")
	if dockerConfigDir != "" {
		envs = append(envs, "DOCKER_CONFIG="+dockerConfigDir)
	}
	cmd.Env = envs

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return apperrors.Wrap(err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return apperrors.Wrap(err)
	}

	if err := cmd.Start(); err != nil {
		return apperrors.Wrap(err)
	}

	var wg sync.WaitGroup
	wg.Go(func() {
		s.streamLogOutput(ctx, data.LogStore, stdout)
	})
	wg.Go(func() {
		s.streamLogOutput(ctx, data.LogStore, stderr)
	})
	wg.Wait()

	if err := cmd.Wait(); err != nil {
		return apperrors.Wrap(err)
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
