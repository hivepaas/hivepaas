package imagebuildserviceimpl

import (
	"context"
	"fmt"
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
	_ context.Context,
	data *imageBuildData,
) error {
	targetPath := filepath.Join(data.CheckoutDir, data.Dockerfile.Path)

	if data.Dockerfile.Content != "" {
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil { //nolint:mnd
			return hperrors.Wrap(err)
		}
		if err := os.WriteFile(targetPath, []byte(data.Dockerfile.Content), 0644); err != nil { //nolint:mnd,gosec
			return hperrors.Wrap(err)
		}
		return nil
	}

	if _, err := os.Stat(targetPath); err != nil {
		if os.IsNotExist(err) {
			return hperrors.NewNotFound("Dockerfile").WithMsgLog("Dockerfile not found at %s", targetPath)
		}
		return hperrors.Wrap(err)
	}

	return nil
}

func (s *service) prepareDockerfileAuto(
	ctx context.Context,
	data *imageBuildData,
) (err error) {
	scanDir := data.CheckoutDir
	if data.Dockerfile.ScanPath != "" {
		scanDir, err = dirInCheckout(data.CheckoutDir, data.Dockerfile.ScanPath)
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

	targetPath := filepath.Join(data.CheckoutDir, data.Dockerfile.Path)
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil { //nolint:mnd
		return hperrors.Wrap(err)
	}
	if err := os.WriteFile(targetPath, []byte(data.Dockerfile.Content), 0644); err != nil { //nolint:mnd,gosec
		return hperrors.Wrap(err)
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

// dirInCheckout is the directory rel names in the checkout, refused when it is
// not one: a path out of it, or through a link the repository holds to a
// directory out of it. The directory becomes the build's context, which the
// image can copy from.
func dirInCheckout(checkoutDir, rel string) (string, error) {
	if !filepath.IsLocal(rel) {
		return "", hperrors.Wrap(hperrors.ErrArgumentInvalid).WithMsgLog("scan path %q leaves the checkout", rel)
	}
	root, err := filepath.EvalSymlinks(checkoutDir)
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	dir, err := filepath.EvalSymlinks(filepath.Join(root, rel))
	if err != nil {
		return "", hperrors.Wrap(hperrors.ErrArgumentInvalid).WithMsgLog("scan path %q: %v", rel, err)
	}
	if inside, err := filepath.Rel(root, dir); err != nil || !filepath.IsLocal(inside) {
		return "", hperrors.Wrap(hperrors.ErrArgumentInvalid).WithMsgLog("scan path %q leaves the checkout", rel)
	}
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		return "", hperrors.Wrap(hperrors.ErrArgumentInvalid).WithMsgLog("scan path %q is not a directory", rel)
	}
	return dir, nil
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
