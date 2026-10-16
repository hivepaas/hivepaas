package imagebuildserviceimpl

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	dockerfile "github.com/hivepaas/dockerfile-generator"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/fileutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
)

func (s *service) prepareDockerfile(
	ctx context.Context,
	data *imageBuildData,
) error {
	data.Dockerfile.Path = gofn.Coalesce(data.Dockerfile.Path, "Dockerfile")

	switch data.Dockerfile.Source {
	case base.DockerfileSourceManual:
		return s.prepareDockerfileManual(ctx, data)
	case base.DockerfileSourceAuto:
		return s.prepareDockerfileAuto(ctx, data)
	}

	return nil
}

func (s *service) prepareDockerfileManual(
	ctx context.Context,
	data *imageBuildData,
) error {
	if data.Dockerfile.Content != "" {
		return writeDockerfile(ctx, data, []byte(data.Dockerfile.Content))
	}

	root, err := os.OpenRoot(data.CheckoutDir)
	if err != nil {
		return hperrors.Wrap(err)
	}
	defer root.Close()
	if _, err := root.Stat(data.Dockerfile.Path); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return hperrors.NewNotFound("Dockerfile").WithMsgLog("Dockerfile not found at %s", data.Dockerfile.Path)
		}
		return refuseDockerfilePath(ctx, data, err)
	}

	return nil
}

// writeDockerfile writes a Dockerfile at its path in the checkout, through
// os.Root: a path that leads out of it - by "..", or through a link the
// repository holds, named as the Dockerfile or on the way to it - is refused,
// never followed to a file of the node.
func writeDockerfile(ctx context.Context, data *imageBuildData, content []byte) error {
	root, err := os.OpenRoot(data.CheckoutDir)
	if err != nil {
		return hperrors.Wrap(err)
	}
	defer root.Close()
	if dir := filepath.Dir(data.Dockerfile.Path); dir != "." {
		if err := root.MkdirAll(dir, 0o755); err != nil { //nolint:mnd
			return refuseDockerfilePath(ctx, data, err)
		}
	}
	if err := root.WriteFile(data.Dockerfile.Path, content, 0o644); err != nil { //nolint:mnd,gosec
		return refuseDockerfilePath(ctx, data, err)
	}
	return nil
}

// readDockerfile reads the Dockerfile at its path in the checkout, through
// os.Root as writeDockerfile writes it.
func readDockerfile(data *imageBuildData) ([]byte, error) {
	root, err := os.OpenRoot(data.CheckoutDir)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	defer root.Close()
	content, err := root.ReadFile(data.Dockerfile.Path)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return content, nil
}

func refuseDockerfilePath(ctx context.Context, data *imageBuildData, err error) error {
	_ = data.LogStore.Add(ctx, tasklog.NewErrFrame(fmt.Sprintf(
		"Dockerfile Path '%s' is not a file of the repository: %v", data.Dockerfile.Path, err), tasklog.TsNow))
	return hperrors.Wrap(hperrors.ErrArgumentInvalid).WithMsgLog("dockerfile path %q: %v", data.Dockerfile.Path, err)
}

func (s *service) prepareDockerfileAuto(
	ctx context.Context,
	data *imageBuildData,
) (err error) {
	scanDir := data.CheckoutDir
	if data.Dockerfile.ScanPath != "" {
		scanDir, err = fileutil.DirInside(data.CheckoutDir, data.Dockerfile.ScanPath)
		if err != nil {
			_ = data.LogStore.Add(ctx, tasklog.NewErrFrame(fmt.Sprintf(
				"Dockerfile Gen Scan Path '%s' is not a directory of the repository", data.Dockerfile.ScanPath),
				tasklog.TsNow))
			return hperrors.Wrap(err)
		}
		// The Dockerfile is written for that directory: its COPY . . is the
		// directory's files, so the directory is the build's context
		data.ContextDir = scanDir
	}

	_ = data.LogStore.Add(ctx, tasklog.NewOutFrame("Start auto-generating Dockerfile from source...",
		tasklog.TsNow))

	df := dockerfile.New(slog.New(&genLogHandler{ctx: ctx, logStore: data.LogStore}))
	contents, r, err := df.Generate(scanDir)
	if err != nil {
		_ = data.LogStore.Add(ctx, tasklog.NewErrFrame("Failed to auto-generate Dockerfile with error: "+
			err.Error(), tasklog.TsNow))
		return hperrors.Wrap(err)
	}

	data.Dockerfile.Content = withBuildVariables(string(contents), data.EnvVars, data.SecretEnvVars)

	if err := writeDockerfile(ctx, data, []byte(data.Dockerfile.Content)); err != nil {
		return err
	}

	_ = data.LogStore.Add(ctx, tasklog.NewOutFrame(
		fmt.Sprintf("Auto-generated Dockerfile for project using '%s'", r.Name()),
		tasklog.TsNow,
	))
	if len(data.EnvVars) > 0 || len(data.SecretEnvVars) > 0 {
		_ = data.LogStore.Add(ctx, tasklog.NewOutFrame(
			"Build-time env variables are given to the Dockerfile's steps: "+
				buildVariableNames(data.EnvVars, data.SecretEnvVars),
			tasklog.TsNow,
		))
	}
	_ = data.LogStore.Add(ctx, tasklog.NewOutFrame(
		"NOTE: The Dockerfile content may contain environment variables for customization. "+
			"To customize, configure them in the app's build environment variables.",
		tasklog.TsNow,
	))
	_ = data.LogStore.Add(ctx, tasklog.NewDebugFrame(data.Dockerfile.Content, tasklog.TsNow))

	return nil
}

// buildVariableNames lists the build's variables for its log, the names alone:
// "A, B, and as secrets: C".
func buildVariableNames(envVars map[string]*string, secretEnvVars map[string]string) string {
	text := strings.Join(slices.Sorted(maps.Keys(envVars)), ", ")
	if len(secretEnvVars) > 0 {
		if text != "" {
			text += "; "
		}
		text += "as secrets: " + strings.Join(slices.Sorted(maps.Keys(secretEnvVars)), ", ")
	}
	return text
}
