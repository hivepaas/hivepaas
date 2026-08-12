package imagebuildserviceimpl

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	dockerfile "github.com/hivepaas/dockerfile-generator"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/reflectutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
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
			return apperrors.Wrap(err)
		}
		if err := os.WriteFile(targetPath, []byte(data.Dockerfile.Content), 0644); err != nil { //nolint:mnd,gosec
			return apperrors.Wrap(err)
		}
		return nil
	}

	if _, err := os.Stat(targetPath); err != nil {
		if os.IsNotExist(err) {
			return apperrors.NewNotFound("Dockerfile").WithMsgLog("Dockerfile not found at %s", targetPath)
		}
		return apperrors.Wrap(err)
	}

	return nil
}

func (s *service) prepareDockerfileAuto(
	ctx context.Context,
	data *imageBuildData,
) (err error) {
	scanDir := data.CheckoutDir
	if data.Dockerfile.ScanPath != "" {
		scanDir = filepath.Join(scanDir, data.Dockerfile.ScanPath)
	}

	s.addStepStartLog(ctx, data, "Start auto-generating Dockerfile from source...")
	defer s.addStepEndLog(ctx, data, timeutil.NowUTC(), err)

	df := dockerfile.New()
	contents, r, err := df.Generate(scanDir)
	if err != nil {
		_ = data.LogStore.Add(ctx, tasklog.NewErrFrame("Failed to auto-generate Dockerfile with error: "+
			err.Error(), tasklog.TsNow))
		return apperrors.Wrap(err)
	}

	data.Dockerfile.Content = reflectutil.UnsafeBytesToStr(contents)

	targetPath := filepath.Join(data.CheckoutDir, data.Dockerfile.Path)
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil { //nolint:mnd
		return apperrors.Wrap(err)
	}
	if err := os.WriteFile(targetPath, contents, 0644); err != nil { //nolint:mnd,gosec
		return apperrors.Wrap(err)
	}

	_ = data.LogStore.Add(ctx, tasklog.NewOutFrame(
		fmt.Sprintf("Auto-generated Dockerfile for project using '%s'", r.Name()),
		tasklog.TsNow,
	))
	_ = data.LogStore.Add(ctx, tasklog.NewOutFrame(
		"NOTE: The Dockerfile content may contain environment variables for customization. "+
			"To customize, configure them in the app's build environment variables.",
		tasklog.TsNow,
	))
	_ = data.LogStore.Add(ctx, tasklog.NewDebugFrame(data.Dockerfile.Content, tasklog.TsNow))

	return nil
}
