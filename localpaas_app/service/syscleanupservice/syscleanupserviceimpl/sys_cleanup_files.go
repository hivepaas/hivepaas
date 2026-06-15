package syscleanupserviceimpl

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/config"
	"github.com/localpaas/localpaas/localpaas_app/service/syscleanupservice"
)

func (s *service) sysCleanupFiles(
	_ context.Context,
	data *sysCleanupData,
) (err error) {
	if !data.SysCleanupSettings.FileCleanup.Enabled {
		return nil
	}

	defer func() {
		if err != nil {
			data.TaskOutput.FileCleanup.Error = err.Error()
		}
	}()

	var errs []error

	// Remove outdated temp files
	errs = append(errs, s.sysCleanupTempFiles(data))

	return errors.Join(errs...)
}

func (s *service) sysCleanupTempFiles(
	data *sysCleanupData,
) (err error) {
	if data.CleanupFilesTemp == syscleanupservice.CleanupFlagFalse {
		return nil
	}

	baseDirs := []string{base.BaseTempDirDefault, filepath.Join(config.Current.AppPath, "tmp")}
	threshold := time.Now().AddDate(0, 0, -3) //nolint:mnd
	if data.CleanupFilesTemp == syscleanupservice.CleanupFlagForce {
		threshold = time.Now()
	}

	for _, baseDir := range baseDirs {
		entries, err := os.ReadDir(baseDir)
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return apperrors.Wrap(err)
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			dirTime, err := time.Parse(time.DateOnly, entry.Name())
			if err != nil {
				continue
			}

			if dirTime.Before(threshold) {
				_ = os.RemoveAll(filepath.Join(baseDir, entry.Name()))
			}
		}
	}

	return nil
}
